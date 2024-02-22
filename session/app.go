package session

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/iooikaak/gateway/model/mysql"
	"io"
	"net/http"
	"strings"

	log "github.com/iooikaak/frame/log"
	"github.com/iooikaak/frame/util"

	"github.com/gomodule/redigo/redis"
	"github.com/iooikaak/gateway/config"
	userpb "github.com/iooikaak/pb/user/pb"
	"github.com/jinzhu/gorm"
)

var (
	errUIDNull        = errors.New("user id is null")
	errTimeStampNull  = errors.New("Auth-Timestamp is null")
	errAuthSignNull   = errors.New("Auth-Sign is null")
	errTokenIncorrect = errors.New("Token Incorrect")
)

var AuthHash string

// 定义平台
type Platform string

const (
	PC  Platform = "pc"
	APP Platform = "app"
)

// TabelName 平台对应的表明
func (p Platform) TabelName() string {
	switch p {
	case PC:
		return "app_user_token_pc"
	default:
		return "app_user_token"
	}
}

// App 加盐认证
type App struct {
	cfg   *config.Config
	dbw   *gorm.DB
	cpool *redis.Pool
}

// NewApp APP加盐认证
func NewApp(cfg *config.Config) Authorization {
	app := new(App)
	app.cpool = util.NewRedisPool(cfg.Redis)
	var err error
	dsnw := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Mysql.User, cfg.Mysql.Psw, cfg.Mysql.Host.Write, cfg.Mysql.DbName)
	app.dbw, err = gorm.Open("mysql", dsnw)
	if err != nil {
		panic(err.Error())
	}
	app.cfg = cfg
	if !cfg.Env.Production() {
		app.dbw.LogMode(true)
	}
	tmpraw := cfg.AuthRaw
	if tmpraw == "" {
		tmpraw = "warhorse2019"
	}
	s := sha1.New()
	io.WriteString(s, tmpraw)
	AuthHash = strings.ToUpper(hex.EncodeToString(s.Sum(nil)))
	log.Infof("AuthHash=%s", AuthHash)
	return app
}

// load token
func (app *App) loadToken(uid string, platform Platform) (string, error) {
	c := app.cpool.Get()
	defer c.Close()
	reply, err := redis.String(c.Do("GET", tokenKey(uid, platform)))
	if err == nil {
		return reply, nil
	}

	var token mysql.Token
	err = app.dbw.Table(platform.TabelName()).Where("user_id = ?", uid).First(&token).Error
	if err != nil {
		return "", err
	}
	if token.Token != "" {
		_, err = c.Do("SETEX", tokenKey(uid, platform), 86400, token.Token)
		if err != nil {
			log.Error(err.Error())
		}
	}
	return token.Token, nil
}

// loadMysqlToken load token from mysql
func (app *App) loadMysqlToken(uid string, platform Platform) (string, error) {
	var token mysql.Token
	err := app.dbw.Table(platform.TabelName()).Where("user_id = ?", uid).First(&token).Error
	if err != nil {
		return "", err
	}
	if token.Token != "" {
		c := app.cpool.Get()
		_, err = c.Do("SETEX", tokenKey(uid, platform), 86400, token.Token)
		if err != nil {
			log.Error(err.Error())
		}
		c.Close()
	}
	return token.Token, nil
}

func signature(uid, token, timestamp string) string {
	k := fmt.Sprintf("warhorse:%s:TOKEN:%s:CT:%s", uid, token, timestamp)
	s := sha1.New()
	io.WriteString(s, k)
	return strings.ToUpper(hex.EncodeToString(s.Sum(nil)))
}

// checkSignature 校验签名是否正确
func (app *App) checkSignature(reqSign, uid, reqTimestamp string, platform Platform) error {
	lToken, err := app.loadToken(uid, platform)
	if err != nil {
		return err
	}
	// 通过缓存的token比对签名，如果不正确再直接从数据库加载token比对签名
	if wSign := signature(uid, lToken, reqTimestamp); wSign != strings.ToUpper(reqSign) {
		lToken, err = app.loadMysqlToken(uid, platform)
		if err != nil {
			return err
		}
		if wSign := signature(uid, lToken, reqTimestamp); wSign != strings.ToUpper(reqSign) {
			return errTokenIncorrect
		}
	}
	return nil
}

// Do 认证用户
func (app *App) Do(r *http.Request) error {
	if authorization := r.Header.Get("Authorization"); authorization != "" && authorization == AuthHash {
		return nil
	}
	var (
		uid       string
		timestamp string
		sign      string
	)
	//用户ID
	if uid = r.Header.Get("userId"); uid == "" {
		return errUIDNull
	}
	//验证通过时间戳
	if timestamp = r.Header.Get("Auth-Timestamp"); timestamp == "" {
		return errTimeStampNull
	}
	//签名
	if sign = r.Header.Get("Auth-Sign"); sign == "" {
		return errAuthSignNull
	}
	//pc还是移动平台
	platform := transfor(r.Header.Get("Auth-Appkey"))
	//校验签名
	return app.checkSignature(sign, uid, timestamp, platform)
}

func tokenKey(uid string, platform Platform) string {
	switch platform {
	case PC:
		return fmt.Sprintf("U:{%s}:AppToken:PC", uid)
	default:
		return fmt.Sprintf("U:{%s}:AppToken", uid)
	}
}

func transfor(appkey string) Platform {
	if userpb.AppPC(appkey) {
		return PC
	} else {
		return APP
	}
}

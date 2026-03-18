package tools

import (
    "fmt"
    "math/rand"
    "strconv"
    "time"

    "github.com/eryajf/go-ldap-admin/config"
    "github.com/patrickmn/go-cache"

    "gopkg.in/gomail.v2"
)

// 验证码放到缓存当中
var VerificationCodeCache = cache.New(24*time.Hour, 48*time.Hour)

// 限制同一账号1分钟只能下发一次验证码
var verificationSendLimiter = cache.New(24*time.Hour, 48*time.Hour)

// VerificationPurpose 区分不同场景的验证码，避免互相污染
type VerificationPurpose string

const (
    PurposeLogin           VerificationPurpose = "login"
    PurposePasswordReset   VerificationPurpose = "password-reset"
    PurposePasswordChange  VerificationPurpose = "password-change"
    PurposeOAuthChallenge  VerificationPurpose = "oauth-challenge"
)

func verificationCacheKey(purpose VerificationPurpose, identity string) string {
    return fmt.Sprintf("%s:%s", purpose, identity)
}

// GenerateVerificationCode 生成并缓存验证码，附带1分钟的发信限流
func GenerateVerificationCode(purpose VerificationPurpose, identity string) (code string, expireAt time.Time, err error) {
    key := verificationCacheKey(purpose, identity)
    if _, found := verificationSendLimiter.Get(key); found {
        return "", time.Time{}, fmt.Errorf("验证码请求过于频繁，请稍后再试")
    }

    rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
    code = fmt.Sprintf("%06v", rnd.Int31n(1000000))
    expireAt = time.Now().Add(5 * time.Minute)

    VerificationCodeCache.Set(key, code, time.Minute*5)
    verificationSendLimiter.Set(key, true, time.Minute)
    return code, expireAt, nil
}

// ValidateVerificationCode 校验验证码，可选是否消费掉
func ValidateVerificationCode(purpose VerificationPurpose, identity, code string, consume bool) bool {
    key := verificationCacheKey(purpose, identity)
    cacheCode, ok := VerificationCodeCache.Get(key)
    if !ok {
        return false
    }
    if cacheCode != code {
        return false
    }
    if consume {
        VerificationCodeCache.Delete(key)
    }
    return true
}

func email(mailTo []string, subject string, body string) error {
	mailConn := map[string]string{
		"user": config.Conf.Email.User,
		"pass": config.Conf.Email.Pass,
		"host": config.Conf.Email.Host,
		"port": config.Conf.Email.Port,
	}
	port, _ := strconv.Atoi(mailConn["port"]) //转换端口类型为int

	newmail := gomail.NewMessage()

	newmail.SetHeader("From", newmail.FormatAddress(mailConn["user"], config.Conf.Email.From))
	newmail.SetHeader("To", mailTo...)    //发送给多个用户
	newmail.SetHeader("Subject", subject) //设置邮件主题
	newmail.SetBody("text/html", body)    //设置邮件正文

	do := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])
	return do.DialAndSend(newmail)
}

func SendMail(sendto []string, pass string) error {
	subject := "重置LDAP密码成功"
	// 邮件正文
	body := fmt.Sprintf("<li><a>更改之后的密码为: %s </a></li>", pass)
	return email(sendto, subject, body)
}

// SendCode 发送验证码
func SendCode(sendto []string) error {
    code, _, err := GenerateVerificationCode(PurposePasswordReset, sendto[0])
    if err != nil {
        return err
    }
    subject := "验证码-重置密码"
    body := fmt.Sprintf(`<div>
        <div>
            尊敬的用户，您好！
        </div>
        <div style="padding: 8px 40px 8px 50px;">
            <p>你本次的验证码为 %s ,为了保证账号安全，验证码有效期为5分钟。请确认为本人操作，切勿向他人泄露，感谢您的理解与使用。</p>
        </div>
        <div>
            <p>此邮箱为系统邮箱，请勿回复。</p>
        </div>
    </div>`, code)
    return email(sendto, subject, body)
}

// SendLoginCode 登录验证码
func SendLoginCode(mail string) (time.Time, error) {
    code, expireAt, err := GenerateVerificationCode(PurposeLogin, mail)
    if err != nil {
        return time.Time{}, err
    }
    subject := "验证码-登录"
    body := fmt.Sprintf(`<div>
        <div>
            尊敬的用户，您好！
        </div>
        <div style="padding: 8px 40px 8px 50px;">
            <p>您正在进行登录操作，本次验证码为 %s ，验证码有效期为5分钟，请勿泄露。</p>
        </div>
        <div>
            <p>此邮箱为系统邮箱，请勿回复。</p>
        </div>
    </div>`, code)
    return expireAt, email([]string{mail}, subject, body)
}

// SendPasswordChangeCode 修改密码验证码
func SendPasswordChangeCode(mail string) (string, time.Time, error) {
    code, expireAt, err := GenerateVerificationCode(PurposePasswordChange, mail)
    if err != nil {
        return "", time.Time{}, err
    }
    subject := "验证码-修改密码"
    body := fmt.Sprintf(`<div>
        <div>
            尊敬的用户，您好！
        </div>
        <div style="padding: 8px 40px 8px 50px;">
            <p>您正在修改账户密码，本次验证码为 %s ，验证码有效期为5分钟，请勿泄露。</p>
        </div>
        <div>
            <p>此邮箱为系统邮箱，请勿回复。</p>
        </div>
    </div>`, code)
    return code, expireAt, email([]string{mail}, subject, body)
}

// SendUserCreationNotification 发送用户创建成功通知邮件
func SendUserCreationNotification(username, nickname, mail, password string) error {
	subject := "LDAP账户创建成功通知"
	// 邮件正文
	body := fmt.Sprintf(`<div>
        <div>
            尊敬的%s，您好！
        </div>
        <div style="padding: 8px 40px 8px 50px;">
            <p>您的LDAP账户已创建成功，以下是您的账户信息：</p>
            <ul>
                <li>用户名：%s</li>
                <li>昵称：%s</li>
                <li>初始密码：%s</li>
            </ul>
            <p style="color: #ff6600;">请妥善保管您的账户信息，建议首次登录后及时修改密码。</p>
        </div>
        <div>
            <p>此邮箱为系统邮箱，请勿回复。</p>
        </div>
    </div>`, nickname, username, nickname, password)
	return email([]string{mail}, subject, body)
}

// SendPasswordResetNotification 发送密码重置成功通知邮件
func SendPasswordResetNotification(username, nickname, mail, newPassword string) error {
	subject := "LDAP密码重置成功通知"
	// 邮件正文
	body := fmt.Sprintf(`<div>
        <div>
            尊敬的%s，您好！
        </div>
        <div style="padding: 8px 40px 8px 50px;">
            <p>您的LDAP账户密码已成功重置，以下是您的新密码信息：</p>
            <ul>
                <li>用户名：%s</li>
                <li>新密码：%s</li>
            </ul>
            <p style="color: #ff6600;">为了您的账户安全，请尽快登录并修改为您自己的密码。</p>
            <p style="color: #ff0000;">请妥善保管您的账户信息，切勿泄露给他人。</p>
        </div>
        <div>
            <p>此邮箱为系统邮箱，请勿回复。</p>
        </div>
    </div>`, nickname, username, newPassword)
	return email([]string{mail}, subject, body)
}

package controller

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/eryajf/go-ldap-admin/public/tools"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zht "github.com/go-playground/validator/v10/translations/zh"
)

var (
	Api           = &ApiController{}
	Group         = &GroupController{}
	Menu          = &MenuController{}
	Role          = &RoleController{}
	User          = &UserController{}
	OperationLog  = &OperationLogController{}
	Base          = &BaseController{}
	FieldRelation = &FieldRelationController{}
	OAuth         = &OAuthController{}

	validate = validator.New()
	trans    ut.Translator
)

// init initializes the validator with Chinese translations and custom validation rules.
func init() {
	uni := ut.New(zh.New())
	trans, _ = uni.GetTranslator("zh")
	_ = zht.RegisterDefaultTranslations(validate, trans)
	_ = validate.RegisterValidation("checkMobile", checkMobile)
}

// checkMobile is a custom validator that checks whether a field value is a valid mobile phone number.
func checkMobile(fl validator.FieldLevel) bool {
	reg := `1\d{10}`
	rgx := regexp.MustCompile(reg)
	return rgx.MatchString(fl.Field().String())
}

// Run binds the request, validates it, and executes the given handler function.
func Run(c *gin.Context, req any, fn func() (any, any)) {
	var err error
	// Bind the incoming request to the struct
	err = c.Bind(req)
	if err != nil {
		tools.Err(c, tools.NewValidatorError(err), nil)
		return
	}
	// Validate the bound struct
	err = validate.Struct(req)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			tools.Err(c, tools.NewValidatorError(fmt.Errorf("%s", err.Translate(trans))), nil)
			return
		}
	}
	data, err1 := fn()
	if err1 != nil {
		tools.Err(c, tools.ReloadErr(err1), data)
		return
	}
	tools.Success(c, data)
}

// Demo is the health check endpoint.
// @Summary Health Check
// @Tags Base Management
// @Produce json
// @Description Health check endpoint that returns a pong response
// @Success 200 {object} response.ResponseBody
// @router /base/ping [get]
func Demo(c *gin.Context) {
	// Health check
	CodeDebug()
	c.JSON(http.StatusOK, tools.H{"code": 200, "msg": "ok", "data": "pong"})
}

// CodeDebug is a placeholder function used for debugging purposes.
func CodeDebug() {
}

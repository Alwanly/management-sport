package binding

import (
	"net/http"
	"reflect"

	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/logger"
	"github.com/Alwanly/management-sport/pkg/middleware"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ContextName = "Binding"

type Binder struct {
	l   *zap.Logger
	ctx *gin.Context
	m   interface{}
}

type Source func(*Binder) error

type ModelBindingError struct {
	Code         int
	ResponseBody wrapper.JSONResult
}

func (e *ModelBindingError) Error() string {
	return "Failed to bind request body"
}

func BindFromBody() Source {
	return func(b *Binder) error {
		if err := b.ctx.ShouldBindJSON(b.m); err != nil {
			b.l.Debug("Error when binding from body", zap.Error(err))
			return err
		}
		return nil
	}
}

func BindFromQuery() Source {
	return func(b *Binder) error {
		if err := b.ctx.ShouldBindQuery(b.m); err != nil {
			b.l.Debug("Error when binding from query string", zap.Error(err))
			return err
		}
		return nil
	}
}

func BindFromParams() Source {
	return func(b *Binder) error {
		if err := b.ctx.ShouldBindUri(b.m); err != nil {
			b.l.Debug("Error when binding from path params", zap.Error(err))
			return err
		}
		return nil
	}
}

func BindFromHeaders() Source {
	return func(b *Binder) error {
		if err := b.ctx.ShouldBindHeader(b.m); err != nil {
			b.l.Debug("Error when binding from request headers", zap.Error(err))
			return err
		}
		return nil
	}
}

func BindFromMultipart() Source {
	return func(b *Binder) error {
		if err := b.ctx.ShouldBind(b.m); err != nil {
			b.l.Debug("Error when binding from multipart form", zap.Error(err))
			return err
		}
		return nil
	}
}

func BindModel(log *zap.Logger, c *gin.Context, m interface{}, sources ...Source) error {
	l := logger.WithID(log, ContextName, "BindModel")

	binder := &Binder{l: l, ctx: c, m: m}

	for _, source := range sources {
		if err := source(binder); err != nil {
			result := wrapper.ResponseFailed(http.StatusBadRequest, contract.StatusCodeBindingFailed, contract.ErrorValidatePayload, nil)
			return &ModelBindingError{Code: result.Code, ResponseBody: result}
		}
	}

	// Set AuthUserData from context
	if authUserValue, exists := c.Get(middleware.LocalTokenKey); exists {
		if authUser, ok := authUserValue.(*middleware.AuthUserData); ok {
			dataField := reflect.Indirect(reflect.ValueOf(m)).FieldByName("AuthUserData")
			if dataField.IsValid() && dataField.CanSet() {
				dataField.Set(reflect.ValueOf(authUser))
			}
		}
	}

	return nil
}

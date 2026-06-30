package common

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/ravenmk2/jungle/internal/apperrors"
)

func TestOKAndFail(t *testing.T) {
	ok := OK(map[string]int{"n": 1})
	if !ok.Success || ok.Data == nil || ok.Error != nil {
		t.Fatalf("OK = %+v", ok)
	}
	fail := Fail(apperrors.New(apperrors.NotFound, "x"))
	if fail.Success || fail.Error == nil || fail.Error.Code != "NOT_FOUND" {
		t.Fatalf("Fail = %+v", fail)
	}
}

func TestMapError(t *testing.T) {
	e := MapError(apperrors.New(apperrors.BuildFailed, "boom"))
	if e.Error.Code != "BUILD_FAILED" {
		t.Fatalf("code = %s", e.Error.Code)
	}
	e2 := MapError(errFoo)
	if e2.Error.Code != "INTERNAL_ERROR" {
		t.Fatalf("unknown err should map to INTERNAL_ERROR, got %s", e2.Error.Code)
	}
}

var errFoo = echo.NewHTTPError(500, "foo")

func TestBindRequired(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	var got struct {
		Workspace string `json:"workspace" validate:"required"`
	}
	if err := Bind(c, &got); err == nil {
		t.Fatal("expected validation error for missing workspace")
	}
}

func TestBindOK(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"workspace":"ws"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	var got struct {
		Workspace string `json:"workspace" validate:"required"`
	}
	if err := Bind(c, &got); err != nil {
		t.Fatal(err)
	}
	if got.Workspace != "ws" {
		t.Fatalf("workspace = %q", got.Workspace)
	}
}

package app

import (
	"log"
	"net/http"
	"os"
	"pvz/internal/config"
	"pvz/internal/database"
	"pvz/internal/handlers"
	"pvz/internal/models"
	"pvz/internal/repository"
	"pvz/internal/services"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type PVZHandlers interface {
	DummyLogin(c echo.Context) error
	RegisterUser(c echo.Context) error
	LoginUser(c echo.Context) error
	CreatePVZ(c echo.Context) error

	CreateReception(c echo.Context) error
	CloseLastReception(c echo.Context) error
	DeleteLastProduct(c echo.Context) error
	CreateProduct(c echo.Context) error

	GetPVZ(c echo.Context) error
}

type App struct {
	Router  *echo.Echo
	Handler PVZHandlers
	Config  config.Config
}

func NewApp(router *echo.Echo, config config.Config) (*App, error) {
	db, err := database.InitDB(config.DB.GetDsn())
	if err != nil {
		return nil, err
	}

	repo := repository.NewRepository(db.DB)
	service := services.NewService(repo)
	handler := handlers.NewHandler(service)

	return &App{Router: router, Handler: handler, Config: config}, nil
}

func (a *App) Start() {
	log.Printf("Starting server at %s", a.Config.GetAddress())
	if err := a.Router.Start(":" + a.Config.GetPort()); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}

func (a *App) RegisterMiddlewares() {
	a.Router.Use(middleware.Recover())
	a.Router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_rfc3339}, method=${method}, uri=${uri}, status=${status}\n",
		Output: os.Stdout,
	}))
}

func (a *App) RegisterRoutes() {
	a.Router.POST("/dummyLogin", a.Handler.DummyLogin)
	a.Router.POST("/register", a.Handler.RegisterUser)
	a.Router.POST("/login", a.Handler.LoginUser)

	jwtMW := echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(os.Getenv("SECRET_KEY")),
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusForbidden, models.Err("access is denied"))
		},
	})

	moderMW := RoleCheckerMW(models.Moderator)
	moderatorsGroup := a.Router.Group("", jwtMW, moderMW)
	moderatorsGroup.POST("/pvz", a.Handler.CreatePVZ)

	employeeMW := RoleCheckerMW(models.Employee)
	employeesGroup := a.Router.Group("", jwtMW, employeeMW)
	employeesGroup.POST("/receptions", a.Handler.CreateReception)
	employeesGroup.POST("/products", a.Handler.CreateProduct)
	employeesGroup.POST("/pvz/:pvzId/close_last_reception", a.Handler.CloseLastReception)
	employeesGroup.POST("/pvz/:pvzId/delete_last_product", a.Handler.DeleteLastProduct)

	moderEmploeeMW := RoleCheckerMW(models.Employee, models.Moderator)
	a.Router.GET("/pvz", a.Handler.GetPVZ, jwtMW, moderEmploeeMW)
}

func RoleCheckerMW(allowedRoles ...models.Role) echo.MiddlewareFunc {
	rolesMap := map[string]struct{}{}
	for _, val := range allowedRoles {
		rolesMap[string(val)] = struct{}{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, ok := c.Get("user").(*jwt.Token)
			if !ok {
				return c.JSON(http.StatusForbidden, models.Err("access is denied"))
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusForbidden, models.Err("access is denied"))
			}

			role, _ := claims["role"].(string)
			if _, ok := rolesMap[role]; !ok {
				return c.JSON(http.StatusForbidden, models.Err("access is denied"))
			}
			return next(c)
		}
	}
}

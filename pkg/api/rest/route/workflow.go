package route

import (
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/labstack/echo/v4"
)

func Workflow(e *echo.Echo) {
	e.POST("/workflow", controller.CreateWorkflow)
	e.GET("/workflow/:id", controller.GetWorkflow)
	e.GET("/workflow", controller.ListWorkflow)
	e.PUT("/workflow/:id", controller.UpdateWorkflow)
	e.POST("/workflow/run/:id", controller.RunWorkflow)
	e.DELETE("/workflow/:id", controller.DeleteWorkflow)
}

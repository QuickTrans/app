package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uploadexpress/app/controllers"
	"github.com/uploadexpress/app/middlewares"
)

func Index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "You successfully reached the upload.express API."})
}

func (a *API) SetupRouter() {
	router := a.Router

	router.Use(middlewares.ErrorMiddleware())
	router.Use(middlewares.ConfigMiddleware(a.Config))
	router.Use(middlewares.StoreMiddleware(a.Database))
	router.Use(middlewares.CorsMiddleware(middlewares.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}))
	router.Use(middlewares.StaticMiddleware("/", middlewares.StaticLocalFile("front", false)))
	router.Use(middlewares.WorkerMiddleware(a.Worker))

	authMiddleware := middlewares.AuthMiddleware()

	v1 := router.Group("/v1")
	{
		v1.GET("/", Index)

		authentication := v1.Group("/auth")
		{
			authController := controllers.NewAuthController()
			authentication.POST("/", authController.Authenticate)
		}

		setup := v1.Group("/setup")
		{
			setupController := controllers.NewSetupController()
			setup.GET("/status", setupController.Status)
			setup.POST("/", setupController.SetupApp)
		}

		uploader := v1.Group("/uploader")
		{
			uploaderController := controllers.NewUploadController()
			uploader.POST("/", uploaderController.Create)
			uploader.PUT("/:upload_id/complete", uploaderController.CompleteUpload)
			uploader.GET("/:upload_id/file/:file_id/upload_url", uploaderController.CreatePreSignedRequest)
			uploader.Use(authMiddleware)
			uploader.GET("/", uploaderController.Index)
			uploader.DELETE("/:upload_id/", uploaderController.DeleteUpload)
		}

		downloader := v1.Group("/downloader")
		{
			downloaderController := controllers.NewDownloaderController()
			downloader.GET("/:download_id", downloaderController.Show)
			downloader.GET("/:download_id/file/:file_id/download_url", downloaderController.GetDownloadLink)
			downloader.GET("/:download_id/zip", downloaderController.DownloadZip)
			downloader.POST("/:download_id/selection/zip", downloaderController.CreateZipWithSelection)
		}

		settings := v1.Group("/settings")
		{
			settingsController := controllers.NewSettingsController()
			settings.GET("/", settingsController.Index)
			uploader.Use(authMiddleware)
			settings.PUT("/", settingsController.Edit)
			settings.POST("/logo/", settingsController.CreateLogo)
			settings.POST("/background/", settingsController.CreateBackground)
			settings.DELETE("/background/:id/", settingsController.DeleteBackground)
			settings.DELETE("/logo/", settingsController.DeleteLogo)

		}
	}

	staticController := controllers.NewStaticController()
	router.LoadHTMLFiles("front/index.html")
	router.NoRoute(staticController.RenderIndex)
}

package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"cudo.iot/traxy_admin/controllers"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"

	cdc "gitlab.com/cudo.core/helper"
)

var glbmemstore cache.Cache

func main() {

	// checkLicense, _ := cdc.CheckLicenseV2("")
	// if !checkLicense {
	// 	fmt.Println("Error License. Your App ID = ", cdc.GetMachineID())
	// 	// os.Exit(0)
	// }
	//---- READ CONFIG JSON ----
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	viper.SetConfigName("app.conf")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
	tmphttpreadheadertimeout, _ := time.ParseDuration(viper.GetString("server.readheadertimeout") + "s")
	tmphttpreadtimeout, _ := time.ParseDuration(viper.GetString("server.readtimeout") + "s")
	tmphttpwritetimeout, _ := time.ParseDuration(viper.GetString("server.writetimeout") + "s")
	tmphttpidletimeout, _ := time.ParseDuration(viper.GetString("server.idletimeout") + "s")
	username := viper.GetString("database.user")
	// password := decryptPw(viper.GetString("database.password"))
	password := viper.GetString("database.password")
	database := viper.GetString("database.name")
	host := viper.GetString("database.host")
	port := viper.GetInt("database.port")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, username, password, database)
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		fmt.Println("Error Connecting DB => ", err)
		os.Exit(0)
	}
	defer db.Close()

	// DBtempo
	databaselog := viper.GetString("database.db_temporary")
	psqlInfolog := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, username, password, databaselog)
	dbtempo, errlog := sqlx.Connect("postgres", psqlInfolog)
	if errlog != nil {
		fmt.Println("Error Connecting DBs => ", errlog)
		os.Exit(0)
	}
	defer dbtempo.Close()

	// scheduler_hours := viper.GetInt("schedule_subscribe")
	maxLifetimelog, _ := time.ParseDuration(viper.GetString("database.max_lifetime_connection") + "s")
	dbtempo.SetMaxIdleConns(viper.GetInt("database.max_idle_dbection"))
	dbtempo.SetConnMaxLifetime(maxLifetimelog)
	dbstempo := cdc.DBStruct{Dbx: dbtempo}
	//---- ROUTING -----
	router := gin.New()
	sessionstore := memstore.NewStore([]byte("SuperDuperRahasia1234567890!@#$%"))
	sessionstore.Options(sessions.Options{
		Path:   "/",     // default /, atau bisa kosongin aja sama aja
		MaxAge: 60 * 30, // dalam satuan detik
	})
	router.Use(sessions.Sessions("SQUASHNMS", sessionstore))
	fmt.Printf("OS: %s\nArchitecture: %s\n", runtime.GOOS, runtime.GOARCH)
	maxLifetime, _ := time.ParseDuration(viper.GetString("database.max_lifetime_connection") + "s")
	db.SetMaxIdleConns(viper.GetInt("database.max_idle_dbection"))
	db.SetConnMaxLifetime(maxLifetime)
	dbs := cdc.DBStruct{Dbx: db}
	Routing(router, dbs, dbstempo)
	pprof.Register(router) //---- RUNNING SERVER WITH PORT -----
	s := &http.Server{
		Addr:              ":" + viper.GetString("server.port"),
		Handler:           router,
		ReadHeaderTimeout: tmphttpreadheadertimeout,
		ReadTimeout:       tmphttpreadtimeout,
		WriteTimeout:      tmphttpwritetimeout,
		IdleTimeout:       tmphttpidletimeout,
		//MaxHeaderBytes:    1 << 20,
	}
	fmt.Println("Server running on port:", viper.GetString("server.port"))
	s.ListenAndServe()
}
func Routing(router *gin.Engine, dbs cdc.DBStruct, dbtempo cdc.DBStruct) {
	// Root static
	router.Static("/assets", "./views/assets/template/template")
	// image product
	router.Static("/product", "./views/assets/image/product")
	// image image_sim
	router.Static("/image_sim", "./views/assets/image/image_sim")
	router.Static("/image_notfound", "./views/assets/image")
	router.Static("/image_sn", "./views/assets/image/image_sn")
	// Load HTML
	router.LoadHTMLGlob("views/templates/*.html")

	// router.POST("/generate_password", func(c *gin.Context) { controllers.ProductView(c) })
	// asset
	router.GET("/", func(c *gin.Context) { controllers.ManageAssetView(c) })
	router.GET("/manage_asset", func(c *gin.Context) { controllers.ManageAssetView(c) })
	// ASSET
	router.POST("/data_asset", func(c *gin.Context) { controllers.DataAsset(c, dbs) })
	router.POST("/add_asset", func(c *gin.Context) { controllers.AddAsset(c, dbs) })
	router.POST("/edit_asset", func(c *gin.Context) { controllers.EditAsset(c, dbs) })
	router.GET("/get_dataasset/:asset_id", func(c *gin.Context) { controllers.GetdataAsset(c, dbs) })
	router.GET("/result_product", func(c *gin.Context) { controllers.ResultDataProduct(c, dbtempo) })
	router.GET("/get_productasset/:product_id", func(c *gin.Context) { controllers.GetDataProduct(c, dbtempo) })
	router.GET("/delete_asset/:asset_id", func(c *gin.Context) { controllers.DeleteAsset(c, dbs) })
	// PACKAGE
	router.GET("/manage_package", func(c *gin.Context) { controllers.ManagePackageView(c) })
	router.POST("/add_package", func(c *gin.Context) { controllers.AddPackage(c, dbtempo) })
	router.POST("/edit_package", func(c *gin.Context) { controllers.EditPackage(c, dbtempo) })
	router.POST("/data_package", func(c *gin.Context) { controllers.DataPackage(c, dbtempo) })
	router.GET("/get_package/:package_id", func(c *gin.Context) { controllers.GetPackage(c, dbtempo) })
	router.GET("/delete_package/:package_id", func(c *gin.Context) { controllers.DeletePackage(c, dbtempo) })
	// product
	router.GET("/manage_product", func(c *gin.Context) { controllers.ManageProductView(c) })
	router.POST("/add_product", func(c *gin.Context) { controllers.AddProduct(c, dbtempo) })
	router.POST("/data_product", func(c *gin.Context) { controllers.DataProduct(c, dbtempo) })
	router.GET("/delete_product/:product_id", func(c *gin.Context) { controllers.DeleteProduct(c, dbtempo) })
	router.GET("/get_product/:product_id", func(c *gin.Context) { controllers.GetProduct(c, dbtempo) })
	router.POST("/edit_product", func(c *gin.Context) { controllers.EditProduct(c, dbtempo) })
	// quality check
	router.GET("/get_dataquality/:asset_id", func(c *gin.Context) { controllers.GetDataQuality(c, dbtempo, dbs) })
	router.POST("/edit_approve", func(c *gin.Context) { controllers.EditQuality(c, dbs) })
	router.POST("/data_quality", func(c *gin.Context) { controllers.DataQuality(c, dbs) })
	router.GET("/manage_quality", func(c *gin.Context) { controllers.ManageQualityView(c) })
}

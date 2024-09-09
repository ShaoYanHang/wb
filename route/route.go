package route

import (
	v1 "app/api/v1"
	"app/middleware"
	"app/utils"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
)

func createMyRender() multitemplate.Renderer {
	p := multitemplate.NewRenderer()
	p.AddFromFiles("admin", "web/admin/dist/index.html")
	return p
}

func InitRouter() {
	gin.SetMode(utils.AppMode)
	r := gin.New()
	// 设置信任网络 []string
	// nil 为不计算，避免性能消耗，上线应当设置
	_ = r.SetTrustedProxies(nil)

	r.HTMLRender = createMyRender()
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.Cors())

	// r.Static("/static", "web/front/dist/static")
	 r.Static("/admin", "web/admin/dist")
	// r.StaticFile("/favicon.ico", "web/front/dist/favicon.ico")

	 // r.GET("/", func(c *gin.Context) {
	 // 	c.HTML(200, "", nil)
	 // })

	  r.GET("/admin", func(c *gin.Context) {
	  	c.HTML(200, "admin", nil)
	  })
	/*
		后台管理路由接口
	*/
	auth := r.Group("api/v1")
	auth.Use(middleware.JwtToken())
	{
		// 用户模块的路由接口
		auth.GET("admin/users", v1.GetUsers)
		auth.PUT("user/:id", v1.EditUser)
		auth.DELETE("user/:id", v1.DeleteUser)
		//修改密码
		auth.PUT("admin/changepw/:id", v1.ChangeUserPassword)

		// // 更新个人设置
		auth.GET("admin/profile/:id", v1.GetProfile)
		auth.PUT("profile/:id", v1.UpdateProfile)
		// 用户信息模块
		auth.POST("user/add", v1.AddUser)
		auth.GET("user/:id", v1.GetUserInfo)
		auth.GET("users", v1.GetUsers)
	}

	router := r.Group("api/v1")
	{
		// // 上传文件
		router.POST("upload1", v1.Upload1)
		router.POST("upload2", v1.Upload2)
		// 展示 FB 文件 没写完
		router.GET("showvcc_record", v1.ShowFile1)
		router.GET("showfb_record", v1.ShowFile2)

		// 展示 FB 每一个账户 每个卡的消耗
		router.POST("showFBDataByaccount", v1.ShowVirtualCardDataByaccount)
		// 展示 FB 每一个账户 每个卡的消耗列表
		router.POST("showFBByaccountList", v1.ShowFBDataByaccountList)
		// 查询所有虚拟卡的余额和总消耗
		router.GET("showVccBalanceAndDeplete", v1.ShowVccBalanceAndDeplete)

		// 需求2
		router.POST("showfb_vccdata", v1.Showfb_vccdata)

		// 打勾 和 备注
		router.POST("updateFBList", v1.UpdateTransactionRecord)
		// 登录控制模块
		router.POST("login", v1.Login)
		router.POST("loginfront", v1.LoginFront)

		// 查看VCCID 即卡号
		router.GET("showVccID", v1.ShowVccID)

		router.GET("showFBID", v1.ShowFBID)

		// 1. 余额 ：开卡 + 充值 + 交易退款 + 交易授权撤销 - 交易授权 - 卡充退
		// 2. 总消耗 ：交易授权
		// router.POST("showVccBalance", v1.ShowVccBalance)
		// router.POST("showVccDeplete", v1.ShowVccDeplete)

		// 3. 月份消耗 ：某个月份的交易授权
		router.POST("showVccDepleteByDate", v1.ShowVccDepleteByDate)

	}
	_ = r.Run(utils.HttpPort)
}

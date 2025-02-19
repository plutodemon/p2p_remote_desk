package auth

import "github.com/gin-gonic/gin"

func Login(c *gin.Context) {
	//var login LoginBody
	//var user models.User
	//err := c.ShouldBindJSON(&login)
	//if err != nil {
	//	c.JSON(400, gin.H{
	//		"message": "error",
	//	})
	//} else {
	//	login.Email = "testdemob@sina.com"
	//	user = FindByEmail(login.Email)
	//
	//	login.Code, _ = google.NewGoogleAuth().GetCode(user.GoogleCode)
	//	login.Password = user.UserPass
	//
	//	code, _ := google.NewGoogleAuth().GetCode(user.GoogleCode)
	//	if login.Code == code && login.Password == user.UserPass {
	//		c.JSON(200, gin.H{
	//			"message": "success",
	//		})
	//	} else {
	//		c.JSON(400, gin.H{
	//			"message": "error",
	//		})
	//	}
	//}
}

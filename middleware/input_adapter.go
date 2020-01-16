package middleware

import (
	"github.com/astaxie/beego"
)

type BeegoController struct {
	Controller beego.Controller
}

func (c *BeegoController) GetQuery(key string) string {
	return c.Controller.Input().Get(key)
}
func (c *BeegoController) GetForm(key string) string {
	return c.Controller.Input().Get(key)
}
func (c *BeegoController) GetHeader(key string) string {
	return c.Controller.Ctx.Input.Header(key)
}
func (c *BeegoController) GetCookie(key string) string {
	return c.Controller.Ctx.Input.Cookie(key)
}
func (c *BeegoController) GetBody() []byte {
	return c.Controller.Ctx.Input.RequestBody
}
func (c *BeegoController) GetPath(key string) string {
	return c.Controller.GetString(":" + key)
}
func (c *BeegoController) GetArray(key string) []string {
	array := make([]string, 0)
	c.Controller.Ctx.Input.Bind(&array, key)
	return array
}

func (c *BeegoController) RenderOutput(ret *RenderStruct) {
	c.Controller.Ctx.Output.Status = ret.Status
	c.Controller.Data["json"] = ret
	c.Controller.ServeJSON()
}

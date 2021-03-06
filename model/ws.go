package model

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"thchat/pkg/logging"
	"thchat/pkg/util"
)

func SearchRoomWS(c *gin.Context) {
	roomName := c.Param("name")
	log.Println("收到js升级请求")
	if room, ok := Rooms[roomName]; ok {
		ServeWS(room.Hub, c)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "房间不存在"})
	}
	log.Println("rooms:", Rooms)

}


func ServeWS(hub *Hub, c *gin.Context){ //开启服务
	//创建连接

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil,)
	if err != nil{
		logging.Error("upgrade e:", err)
		return
	}

	username := util.GetSessionUsername(c)
	logging.Info("有人来访房间", c.Param("name"), ";客人：", username)
	//创建一个客户端
	client := &Client{hub: hub, conn:conn, send: make(chan []byte, 256), user:&User{Username:username}}
	client.hub.register <- client


	go client.PumpToHub()
	go client.ReadFromHub()
}
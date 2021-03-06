package model

import (
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"net/http"
	"thchat/pkg/logging"
	"time"
)

const (
	//客户端（pong）等待ping的最长时间
	pongWait = 60 * time.Second

	//hub发ping频率
	pingPeriod = 50 * time.Second
)

//声明一个upgrader（把http升级成ws（用于创建ws.conn
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	hub 	*Hub
	conn 	*websocket.Conn
	send 	chan []byte
	user 	*User
}

func (c *Client)SetPong(){

	//超过该时间就断开连接
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil{
		logging.Error("SetDDL failed:",err)
	}

	//听ping（不用发pong
	c.conn.SetPongHandler(func(string)error{
		err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil{ return err}
		return nil
	})
}

//把消息泵到hub里
func (c *Client)PumpToHub(){
	defer func(){
		c.hub.unregister <- c	//注销
		c.conn.Close()
	}()
	c.SetPong()
	for{
		_, msg, err := c.conn.ReadMessage()
		if err != nil{
			logging.Info("One connection closed.")
			break
		}
		logging.Info(fmt.Sprintf("%s: %s", c.user.Username, string(msg)))
		msg = []byte(c.user.Username + ": " + string(msg))
		c.hub.broadcast <- msg

	}
}

//从hub写到连接里
func (c *Client)ReadFromHub(){
	ticker := time.NewTicker(pingPeriod)
	for{
		select {

			case msg,ok := <-c.send:
				if !ok{
				//说明连接已经关了
				return
			}
				//写入conn的writer
				w, err := c.conn.NextWriter(websocket.TextMessage)
				if err != nil{
					logging.Error("next writer:",err)
				}
				//过滤输入
				template.HTMLEscape(w, msg)
				err = w.Close()
				if err != nil{
					logging.Error("w close failed:",err)
					}

			//保持心跳(hub发ping到客户端
			case <-ticker.C:
				err := c.conn.WriteMessage(websocket.PingMessage, nil)
				if err != nil{
					logging.Error("tick e:", err)
					}
		}
	}
}


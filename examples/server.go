package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wazsmwazsm/longpoll"
	"net/http"
	"strconv"
	"time"
)

// Server for longpoll
type Server struct {
	httpSrv *http.Server
	lpSrv   longpoll.ILongPoll
}

// NewServer create longpoll server
func NewServer(host string, port int) *Server {

	lpSrv := longpoll.NewLongPoll()
	srv := &Server{
		lpSrv: lpSrv,
	}
	mux := gin.New()
	mux.GET("/sub", srv.longpollHandler)
	mux.POST("/pub", srv.publish)
	srv.httpSrv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: mux,
	}
	return srv
}

// Run server
func (s *Server) Run() error {
	if err := s.httpSrv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (s *Server) longpollHandler(c *gin.Context) {

	topic := c.Query("topic")
	id := c.Query("id")
	timeout := c.Query("timeout")

	timeoutInt, _ := strconv.ParseInt(timeout, 10, 64)

	sub := s.lpSrv.Subscribe(topic, id, time.Duration(timeoutInt)*time.Second)

	event, err := sub.WaitEvent()
	if err != nil {
		if err == longpoll.ErrTimeout {
			c.JSON(http.StatusOK, gin.H{
				"errMsg": "Long polling timeout",
				"data":   struct{}{},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"errMsg": err.Error(),
			"data":   struct{}{},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"errMsg": "",
		"data":   event,
	})
}

type pubData struct {
	Topic string `json:"topic"`
	Data  string `json:"data"`
}

func (s *Server) publish(c *gin.Context) {

	req := new(pubData)

	if err := c.BindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errMsg": err.Error(),
			"data":   struct{}{},
		})
		return
	}
	s.lpSrv.Publish(req.Topic, req.Data)

}
func main() {
	s := NewServer("0.0.0.0", 8080)
	if err := s.Run(); err != nil {
		panic(err)
	}
}

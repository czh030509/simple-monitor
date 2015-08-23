package main

import (
	"fmt"
	"sync"
	"errors"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"

	"math/rand"
	"time"
	"strconv"
	"github.com/czh030509/simple-monitor/simple-monitor/util")

func main() {
	fmt.Println("Monitor Service starting...")
	r := gin.Default()

	r.Use(static.Serve("/", static.LocalFile("static", true)))

	r.GET("/api/GetServerGroups", func(c *gin.Context) {
		groups, err := ReadData()
		if err == nil {
			groups = FillPublishInfo(groups) // Get publish info
			c.JSON(200, groups)
		} else {
			c.String(500, "ReadData error: "+err.Error())
		}
	})

	r.POST("/api/SaveServer", func(c *gin.Context) {
		var s Server
		if err := c.BindJSON(&s); err == nil {
			err2 := UpdateData(s)

			if  err2 == nil {
				c.String(200, "OK")
			} else {
				c.String(500, "BindJSON error: "+ err2.Error())
			}
		} else {
			c.String(500, "BindJSON error: "+err.Error())
		}
	})

	fmt.Println("Monitor Service started!")
	r.Run(":1122")
}

type ServerGroup struct {
	GroupName string
	Servers   []*Server
}

const (
	Status_Available = "Available"
	Status_Using     = "Using"
)

type Server struct {
	ServerName   string
	Status       string
	Username     string
	UsingEndTime string
	Publishes    []ServerPublish
	GroupName    string
}

type ServerPublish struct {
	Success bool
	Link    string
}

func FillPublishInfo(groups []ServerGroup) []ServerGroup {
	for _, g := range groups {
		for _, s := range g.Servers {
			// Get server's publish info
			// Here just add some random
			s.Publishes = []ServerPublish{}
			pubCount := rand.Intn(4)
			i := 0
			for i <= pubCount {
				sp := ServerPublish{
					Link: "http://www.baidu.com",
				}
				if rand.Float64() > 0.5 {
					sp.Success = true
				}
				s.Publishes = append(s.Publishes, sp)
				i++
			}
		}
	}
	return groups
}

const DataFileName = "ServerGroups.json"

var dataLock = &sync.Mutex{}

func UpdateData(newData Server) error {
	dataLock.Lock()
	defer dataLock.Unlock()

	groups, err := ReadData()
	if err != nil {
		return err
	}
	for _, g := range groups {
		for _, s := range g.Servers {
			if newData.GroupName == g.GroupName && newData.ServerName == s.ServerName {

				fmt.Println(s.Status + "...." +  newData.Status)
				if s.Status == Status_Using && newData.Status == Status_Using {
					fmt.Println("已被占用！")
					return errors.New(s.ServerName + " 已被占用！")
				}

				if s.Status == Status_Available && newData.Status == Status_Available {
					fmt.Println("已被释放！")
					return errors.New(s.ServerName + " 已被释放！")
				}

				if newData.Status == Status_Using {
					h,err := strconv.Atoi(newData.UsingEndTime)

					if err != nil {
						return errors.New("输入使用时间不正确（小时）！")
					}

					s.UsingEndTime = time.Now().Add(time.Duration(h) * time.Hour).Format("2006-01-02 15:04:05")
				} else {
					s.UsingEndTime = newData.UsingEndTime
				}

				s.Status = newData.Status
				s.Username = newData.Username
			}
		}
	}
	jsonStr, err := util.EncodeToJSON(groups)
	if err != nil {
		return err
	}
	return util.WriteFileString(DataFileName, jsonStr)
}

func ReadData() ([]ServerGroup, error) {

	fileContent, err := util.ReadFileString(DataFileName)
	if err != nil {
		return nil, err
	}
	var groups []ServerGroup
	if err = util.DecodeJSON(fileContent, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

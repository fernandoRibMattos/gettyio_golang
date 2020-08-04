package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ========== Data

type Data struct {
	Id    bson.ObjectId `form:"id" bson:"_id,omitempty" json:"id"`
	Name  string        `form:"name" bson:"name" json:"name"`
	Idade int           `form:"idade" bson:"idade" json:"idade"`
}

// ========== MongoDB

type MongoDB struct {
	Host             string
	Port             string
	Addrs            string
	Database         string
	EventTTLAfterEnd time.Duration
	StdEventTTL      time.Duration
	Info             *mgo.DialInfo
	Session          *mgo.Session
}

func (mongo *MongoDB) SetDefault() { // {{{
	mongo.Host = "localhost"
	mongo.Addrs = "localhost:27017"
	mongo.Database = "gettyio_01"
	mongo.EventTTLAfterEnd = 1 * time.Second
	mongo.StdEventTTL = 20 * time.Minute
	mongo.Info = &mgo.DialInfo{
		Addrs:    []string{mongo.Addrs},
		Timeout:  60 * time.Second,
		Database: mongo.Database,
	}
} // }}}

func (mongo *MongoDB) Drop() (err error) { // {{{
	session := mongo.Session.Clone()
	defer session.Close()

	err = session.DB(mongo.Database).DropDatabase()
	if err != nil {
		return err
	}
	return nil
} // }}}

func (mongo *MongoDB) Init() (err error) { // {{{
	err = mongo.Drop()
	if err != nil {
		fmt.Printf("\n drop database error: %v\n", err)
	}

	data := Data{}
	err = mongo.PostData(&data)

	return err
} // }}}

func (mongo *MongoDB) SetSession() (err error) {
	mongo.Session, err = mgo.DialWithInfo(mongo.Info)
	if err != nil {
		mongo.Session, err = mgo.Dial(mongo.Host)
		if err != nil {
			return err
		}
	}
	return err
}

// ========== model

func (mongo *MongoDB) GetData() (dates []Data, err error) { // {{{
	session := mongo.Session.Clone()
	defer session.Close()

	err = session.DB(mongo.Database).C("Data").Find(bson.M{}).All(&dates)
	return dates, err
} // }}}

func (mongo *MongoDB) PostData(data *Data) (err error) { // {{{
	session := mongo.Session.Clone()
	defer session.Close()

	err = session.DB(mongo.Database).C("Data").Insert(&data)
	return err
} // }}}

// ========== controller

func getData(c *gin.Context) { // {{{
	mongo, ok := c.Keys["mongo"].(*MongoDB)
	if !ok {
		c.JSON(400, gin.H{"message": "can't reach db", "body": nil})
	}

	data, err := mongo.GetData()
	// fmt.Printf("\ndata: %v, ok: %v\n", data, ok)
	if err != nil {
		c.JSON(400, gin.H{"message": "can't get data from database", "body": nil})
	} else {
		c.JSON(200, gin.H{"message": "get data sucess", "body": data})
	}
} // }}}

func postData(c *gin.Context) { // {{{
	mongo, ok := c.Keys["mongo"].(*MongoDB)
	if !ok {
		c.JSON(400, gin.H{"message": "can't connect to db", "body": nil})
	}
	var req Data
	err := c.Bind(&req)
	if err != nil {
		c.JSON(400, gin.H{"message": "Incorrect data", "body": nil})
		return
	} else {
		err := mongo.PostData(&req)
		if err != nil {
			c.JSON(400, gin.H{"message": "error post to db", "body": nil})
		}
		c.JSON(200, gin.H{"message": "post data sucess", "body": req})
	}
} // }}}

// ========== middleware

func MiddleDB(mongo *MongoDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := mongo.SetSession()
		if err != nil {
			c.Abort()
		} else {
			c.Set("mongo", mongo)
			c.Next()
		}
	}
}

// ========== start router

func SetupRouter() *gin.Engine {
	mongo := MongoDB{}
	mongo.SetDefault()

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(MiddleDB(&mongo))

	router.GET("/data", getData)
	router.POST("/data", postData)
	return router
}

func main() {
	router := SetupRouter()
	router.Run()
}

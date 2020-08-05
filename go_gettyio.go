package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ========== Tabela CLIENTE

type Cliente struct {
	Id    bson.ObjectId `form:"id" bson:"_id,omitempty" json:"id"`
	Name  string        `form:"name" bson:"name" json:"name"`
	Idade int           `form:"idade" bson:"idade" json:"idade"`
}

// ========== MongoDB - Estrutura

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

// ========== MongoDB - Configuração

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

// ========== MongoDB -

func (mongo *MongoDB) Drop() (err error) { // {{{
	session := mongo.Session.Clone()
	defer session.Close()

	err = session.DB(mongo.Database).DropDatabase()
	if err != nil {
		return err
	}
	return nil
} // }}}

// ========== MongoDB -

func (mongo *MongoDB) Init() (err error) { // {{{
	err = mongo.Drop()
	if err != nil {
		fmt.Printf("\n drop database error: %v\n", err)
	}

	cliente := Cliente{}
	err = mongo.PostCliente(&cliente)

	return err
} // }}}

// ========== MongoDB -

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

// ========== MODEL

// ========== Consulta - Cliente

func (mongo *MongoDB) GetCliente() (clientes []Cliente, err error) { // {{{
	session := mongo.Session.Clone()
	defer session.Close()

	err = session.DB(mongo.Database).C("Cliente").Find(bson.M{}).All(&clientes)
	return clientes, err
} // }}}

// ========== Insere - Cliente

func (mongo *MongoDB) PostCliente(cliente *Cliente) (err error) { // {{{
	session := mongo.Session.Clone()
	defer session.Close()

	err = session.DB(mongo.Database).C("Cliente").Insert(&cliente)
	return err
} // }}}

// ========== Altera - Cliente

func (mongo *MongoDB) PutCliente(cliente *Cliente) (err error) { // {{{
	session := mongo.Session.Clone()
	defer session.Close()

	err = session.DB(mongo.Database).C("Cliente").UpdateId(cliente.Id, bson.M{"$set": &cliente})
	return err
} // }}}

// ========== Delete - Cliente

func (mongo *MongoDB) DeleteCliente(cliente *Cliente) (err error) { // {{{
	session := mongo.Session.Clone()
	defer session.Close()

	err = session.DB(mongo.Database).C("Cliente").RemoveId(cliente.Id)
	return err
} // }}}

// ========== controller

func getCliente(c *gin.Context) { // {{{
	mongo, ok := c.Keys["mongo"].(*MongoDB)
	if !ok {
		c.JSON(400, gin.H{"message": "can't reach db", "body": nil})
	}

	data, err := mongo.GetCliente()
	// fmt.Printf("\ndata: %v, ok: %v\n", data, ok)
	if err != nil {
		c.JSON(400, gin.H{"message": "can't get data from database", "body": nil})
	} else {
		c.JSON(200, gin.H{"message": "get data sucess", "body": data})
	}
} // }}}

func postCliente(c *gin.Context) { // {{{
	mongo, ok := c.Keys["mongo"].(*MongoDB)
	if !ok {
		c.JSON(400, gin.H{"message": "can't connect to db", "body": nil})
	}
	var req Cliente
	err := c.Bind(&req)
	if err != nil {
		c.JSON(400, gin.H{"message": "Incorrect data", "body": nil})
		return
	} else {
		err := mongo.PostCliente(&req)
		if err != nil {
			c.JSON(400, gin.H{"message": "error post to db", "body": nil})
		}
		c.JSON(200, gin.H{"message": "post data sucess", "body": req})
	}
} // }}}

func putCliente(c *gin.Context) { // {{{
	mongo, ok := c.Keys["mongo"].(*MongoDB)
	if !ok {
		c.JSON(400, gin.H{"message": "can't connect to db", "body": nil})
	}
	var req Cliente
	err := c.Bind(&req)
	if err != nil {
		c.JSON(400, gin.H{"message": "Incorrect data", "body": nil})
		return
	} else {
		err := mongo.PutCliente(&req)
		if err != nil {
			c.JSON(400, gin.H{"message": "error put to db", "body": nil})
		}
		c.JSON(200, gin.H{"message": "put data sucess", "body": req})
	}
} // }}}

func deleteCliente(c *gin.Context) { // {{{
	mongo, ok := c.Keys["mongo"].(*MongoDB)
	if !ok {
		c.JSON(400, gin.H{"message": "can't connect to db", "body": nil})
	}
	var req Cliente
	err := c.Bind(&req)
	if err != nil {
		c.JSON(400, gin.H{"message": "Incorrect data", "body": nil})
		return
	} else {
		err := mongo.DeleteCliente(&req)
		if err != nil {
			c.JSON(400, gin.H{"message": "error delete to db", "body": nil})
		}
		c.JSON(200, gin.H{"message": "delete data sucess", "body": req})
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

	router.GET("/cliente", getCliente)
	router.POST("/cliente", postCliente)
	router.PUT("/cliente", putCliente)
	router.DELETE("/cliente", deleteCliente)
	return router
}

func main() {
	router := SetupRouter()
	router.Run()
}

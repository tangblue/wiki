package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/spec"
)

type JWTToken struct {
	Token string `json:"token"`
}

type User struct {
	ID   string `json:"id" description:"identifier of the user" default:"1"`
	Name string `json:"name" description:"name of the user" default:"john"`
	Age  int    `json:"age" description:"age of the user" default:"21"`
}

type UserResource struct {
	// normally one would use DAO (data access object)
	users map[string]User
}

func basicAuthenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// usr/pwd = admin/admin
	u, p, ok := req.Request.BasicAuth()
	if !ok || u != "admin" || p != "admin" {
		resp.AddHeader("WWW-Authenticate", "Basic realm=Protected Area")
		resp.WriteErrorString(401, "401: Not Authorized")
		return
	}
	chain.ProcessFilter(req, resp)
}

const secret = "secret"

func JWTAuthenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	ah := req.HeaderParameter("authorization")
	bearerToken := strings.Fields(ah)
	if len(bearerToken) != 2 {
		resp.WriteErrorString(401, "401: Not Authorized")
		return
	}

	token, error := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}
		return []byte(secret), nil
	})
	if error != nil || !token.Valid {
		resp.WriteErrorString(401, "401: Not Authorized")
		return
	}
	log.Printf("Claims: %v", token.Claims)

	chain.ProcessFilter(req, resp)
}

func (u UserResource) WebService() *restful.WebService {
	printPath := func(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
		log.Printf("Path: %v", request.Request.URL.Path)
		chain.ProcessFilter(request, response)
	}

	ws := new(restful.WebService)
	ws.Path("/users").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML).
		Filter(printPath)

	tags := []string{"users"}
	tagUsers := func(b *restful.RouteBuilder) {
		b.Metadata(restfulspec.KeyOpenAPITags, tags)
	}

	uid := ws.PathParameter("userID", "identifier of the user").DataType("string").DefaultValue("1")
	ws.Route(ws.GET("/").To(u.findAllUsers).
		Doc("get all users").
		Do(tagUsers).
		Filter(JWTAuthenticate).
		Writes([]User{}).
		Returns(200, "OK", []User{}))

	ws.Route(ws.GET("/authenticate").To(u.createToken).
		Doc("authenticate").
		Do(tagUsers).
		Filter(basicAuthenticate).
		Returns(200, "OK", JWTToken{}))

	ws.Route(ws.GET("/{userID}").To(u.findUser).
		Doc("get a user").
		Do(tagUsers).
		Param(uid).
		Writes(User{}).
		Returns(200, "OK", User{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.PUT("/{userID}").To(u.updateUser).
		Doc("update a user").
		Do(tagUsers).
		Param(uid).
		Reads(User{}).
		Writes(User{}).
		Returns(200, "OK", User{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.PUT("").To(u.createUser).
		Doc("create a user").
		Do(tagUsers).
		Reads(User{}))

	ws.Route(ws.DELETE("/{userID}").To(u.removeUser).
		Doc("delete a user").
		Do(tagUsers).
		Param(uid))

	return ws
}

func (u UserResource) createToken(request *restful.Request, response *restful.Response) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "admin",
	})
	tokenString, error := token.SignedString([]byte(secret))
	if error != nil {
		response.WriteErrorString(http.StatusNotFound, "User could not be found.")
	} else {
		response.WriteEntity(JWTToken{Token: tokenString})
	}
}

func (u UserResource) findAllUsers(request *restful.Request, response *restful.Response) {
	list := []User{}
	for _, each := range u.users {
		list = append(list, each)
	}
	response.WriteEntity(list)
}

func (u UserResource) findUser(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("userID")
	if usr, ok := u.users[id]; !ok {
		response.WriteErrorString(http.StatusNotFound, "User could not be found.")
	} else {
		response.WriteEntity(usr)
	}
}

func (u *UserResource) updateUser(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("userID")
	usr, ok := u.users[id]
	if !ok {
		response.WriteErrorString(http.StatusNotFound, "User could not be found.")
		return
	}

	if err := request.ReadEntity(&usr); err == nil {
		usr.ID = id
		u.users[id] = usr
		response.WriteEntity(usr)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func (u *UserResource) createUser(request *restful.Request, response *restful.Response) {
	usr := User{}
	if err := request.ReadEntity(&usr); err == nil {
		u.users[usr.ID] = usr
		response.WriteHeaderAndEntity(http.StatusCreated, usr)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func (u *UserResource) removeUser(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("userID")
	delete(u.users, id)
}

func main() {
	u := UserResource{map[string]User{}}
	restful.DefaultContainer.Add(u.WebService())

	swaggerJson := "/apidocs.json"
	config := restfulspec.Config{
		WebServices: restful.RegisteredWebServices(),
		APIPath:     swaggerJson,
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	basePath := "/apidocs/"
	http.Handle(basePath, http.StripPrefix(basePath, http.FileServer(http.Dir("./vendor/github.com/swagger-api/swagger-ui/dist"))))

	// Optionally, you may need to enable CORS for the UI to work.
	cors := restful.CrossOriginResourceSharing{
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		CookiesAllowed: false,
		Container:      restful.DefaultContainer}
	restful.DefaultContainer.Filter(cors.Filter)

	url := "http://localhost:8080"
	swaggerJson = url + swaggerJson
	log.Printf("Get the API: " + swaggerJson)
	log.Printf("Swagger UI : " + url + basePath + "?url=" + swaggerJson)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "UserService",
			Description: "Resource for managing Users",
			Contact: &spec.ContactInfo{
				Name:  "user",
				Email: "user@example.com",
				URL:   "http://example.com",
			},
			License: &spec.License{
				Name: "MIT",
				URL:  "http://mit.org",
			},
			Version: "1.0.0",
		},
	}
	swo.Tags = []spec.Tag{spec.Tag{TagProps: spec.TagProps{
		Name:        "users",
		Description: "Managing users"}}}
}

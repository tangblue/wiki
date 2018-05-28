package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-openapi/spec"
	"github.com/tangblue/goapi/restful"
	"github.com/tangblue/goapi/restfulspec"
)

type LoginInfo struct {
	Name     string `json:"name" description:"user name"`
	Password string `json:"password" description:"password"`
}

type JWTToken struct {
	Token string `json:"token" description:"JWT token"`
}

type Auth struct {
	secret          string
	hpAuthorization *restful.Parameter
}

func (a *Auth) WebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/login").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("").To(a.createToken).
		Doc("login").
		Metadata(restfulspec.KeyOpenAPITags, []string{"authentication"}).
		Reads(LoginInfo{}).
		Returns(http.StatusOK, "OK", JWTToken{}).
		Returns(http.StatusInternalServerError, "Internal Server Error", nil).
		Returns(http.StatusUnprocessableEntity, "Bad user name or password", nil))

	return ws
}

func (a *Auth) basicAuthenticate(req *restful.Request, resp *restful.Response, next func(*restful.Request, *restful.Response)) {
	// usr/pwd = admin/admin
	u, p, ok := req.Request.BasicAuth()
	if !ok || u != "admin" || p != "admin" {
		resp.AddHeader("WWW-Authenticate", "Basic realm=Protected Area")
		resp.WriteErrorString(http.StatusUnauthorized, "401: Not Authorized")
		return
	}
	next(req, resp)
}

func (a *Auth) createToken(req *restful.Request, resp *restful.Response) {
	li := LoginInfo{}
	if err := req.ReadEntity(&li); err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": li.Name,
	})
	tokenString, err := token.SignedString([]byte(a.secret))
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
	}
	resp.WriteEntity(JWTToken{Token: tokenString})
}

func (a *Auth) JWTAuthenticate(req *restful.Request, resp *restful.Response, next func(*restful.Request, *restful.Response)) {
	ah := req.GetParameter(a.hpAuthorization)
	bt := strings.Fields(ah)
	if len(bt) != 2 || !strings.EqualFold(bt[0], "bearer") {
		resp.WriteErrorString(http.StatusUnauthorized, "401: Not Authorized")
		return
	}

	token, err := jwt.Parse(bt[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}
		return []byte(a.secret), nil
	})
	if err != nil || !token.Valid {
		resp.WriteErrorString(http.StatusUnauthorized, "401: Not Authorized")
		return
	}
	log.Printf("Claims: %v", token.Claims)

	next(req, resp)
}

type User struct {
	ID   int    `json:"id" description:"identifier of the user" default:"1"`
	Name string `json:"name" description:"name of the user" default:"john"`
	Age  int    `json:"age" description:"age of the user" default:"21"`
}

type UserResource struct {
	// normally one would use DAO (data access object)
	users map[int]User
	ppUID *restful.Parameter

	auth *Auth
}

func (u UserResource) WebService() *restful.WebService {
	printPath := func(req *restful.Request, resp *restful.Response, next func(*restful.Request, *restful.Response)) {
		log.Printf("Path: %v", req.Request.URL.Path)
		next(req, resp)
	}
	tagUsers := func(b *restful.RouteBuilder) {
		b.Metadata(restfulspec.KeyOpenAPITags, []string{"users"})
	}
	addAuth := func(b *restful.RouteBuilder) {
		b.Filter(u.auth.JWTAuthenticate).Param(u.auth.hpAuthorization)
	}

	ws := new(restful.WebService)
	ws.Path("/users").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML).
		Filter(printPath)

	ws.Route(ws.GET("/").To(u.findAllUsers).
		Doc("get all users").
		Do(tagUsers).
		Filter(u.auth.basicAuthenticate).
		Returns(http.StatusOK, "OK", []User{}).
		Returns(http.StatusUnauthorized, "Not Authorized", nil))

	ws.Route(ws.PUT("").To(u.createUser).
		Doc("create a user").
		Do(tagUsers, addAuth).
		Reads(User{}).
		Returns(http.StatusCreated, "Created", User{}).
		Returns(http.StatusUnauthorized, "Not Authorized", nil))

	ws.Route(ws.GET("/{%s}", u.ppUID).To(u.findUser).
		Doc("get a user").
		Do(tagUsers).
		Returns(http.StatusOK, "OK", User{}).
		Returns(http.StatusNotFound, "Not Found", nil))

	ws.Route(ws.PUT("/{%s}", u.ppUID).To(u.updateUser).
		Doc("update a user").
		Do(tagUsers, addAuth).
		Reads(User{}).
		Returns(http.StatusOK, "OK", User{}).
		Returns(http.StatusUnauthorized, "Not Authorized", nil).
		Returns(http.StatusNotFound, "Not Found", nil))

	ws.Route(ws.DELETE("/{%s}", u.ppUID).To(u.removeUser).
		Doc("delete a user").
		Do(tagUsers, addAuth).
		Returns(http.StatusNoContent, "No Content", nil).
		Returns(http.StatusUnauthorized, "Not Authorized", nil).
		Returns(http.StatusNotFound, "Not Found", nil))

	return ws
}

func (u UserResource) findAllUsers(req *restful.Request, resp *restful.Response) {
	list := []User{}
	for _, each := range u.users {
		list = append(list, each)
	}
	resp.WriteEntity(list)
}

func (u UserResource) findUser(req *restful.Request, resp *restful.Response) {
	id, err := strconv.Atoi(req.GetParameter(u.ppUID))
	if err != nil {
		resp.WriteErrorString(http.StatusBadRequest, "User ID is invalid.")
		return
	}

	if usr, ok := u.users[id]; !ok {
		resp.WriteErrorString(http.StatusNotFound, "User could not be found.")
	} else {
		resp.WriteEntity(usr)
	}
}

func (u *UserResource) updateUser(req *restful.Request, resp *restful.Response) {
	id, err := strconv.Atoi(req.GetParameter(u.ppUID))
	if err != nil {
		resp.WriteErrorString(http.StatusBadRequest, "User ID is invalid.")
		return
	}

	usr, ok := u.users[id]
	if !ok {
		resp.WriteErrorString(http.StatusNotFound, "User could not be found.")
		return
	}

	if err := req.ReadEntity(&usr); err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	usr.ID = id
	u.users[id] = usr
	resp.WriteEntity(usr)
}

func (u *UserResource) createUser(req *restful.Request, resp *restful.Response) {
	usr := User{}
	if err := req.ReadEntity(&usr); err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}
	u.users[usr.ID] = usr
	resp.WriteHeaderAndEntity(http.StatusCreated, usr)
}

func (u *UserResource) removeUser(req *restful.Request, resp *restful.Response) {
	id, err := strconv.Atoi(req.GetParameter(u.ppUID))
	if err != nil {
		resp.WriteErrorString(http.StatusBadRequest, "User ID is invalid.")
		return
	}
	delete(u.users, id)
	resp.WriteHeader(http.StatusNoContent)
}

func main() {
	auth := &Auth{
		secret:          "secret",
		hpAuthorization: restful.HeaderParameter("authorization", "JWT in authorization header"),
	}
	restful.DefaultContainer.Add(auth.WebService())

	u := UserResource{
		users: map[int]User{},
		ppUID: restful.PathParameter("userID", "identifier of the user").DataType("int").Regex("\\d+").DefaultValue("1"),
		auth:  auth,
	}
	restful.DefaultContainer.Add(u.WebService())

	swaggerJson := "/apidocs.json"
	config := restfulspec.Config{
		WebServices: restful.RegisteredWebServices(),
		APIPath:     swaggerJson,
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	basePath := "/apidocs/"
	http.Handle(basePath, http.StripPrefix(basePath, http.FileServer(http.Dir("./swagger-ui/dist"))))

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
	swo.Tags = []spec.Tag{
		spec.Tag{
			TagProps: spec.TagProps{
				Name:        "authentication",
				Description: "Authentication",
			},
		},
		spec.Tag{
			TagProps: spec.TagProps{
				Name:        "users",
				Description: "Managing users",
			},
		},
	}
}

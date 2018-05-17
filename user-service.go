package main

import (
	"log"
	"net/http"

	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/spec"
)

type User struct {
	ID   string `json:"id" description:"identifier of the user" default:"1"`
	Name string `json:"name" description:"name of the user" default:"john"`
	Age  int    `json:"age" description:"age of the user" default:"21"`
}

type UserResource struct {
	// normally one would use DAO (data access object)
	users map[string]User
}

func (u UserResource) WebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/users").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"users"}
	tagUsers := func(b *restful.RouteBuilder) {
		b.Metadata(restfulspec.KeyOpenAPITags, tags)
	}

	uid := ws.PathParameter("id", "identifier of the user").DataType("string").DefaultValue("1")
	ws.Route(ws.GET("/").To(u.findAllUsers).
		Doc("get all users").
		Do(tagUsers).
		Writes([]User{}).
		Returns(200, "OK", []User{}))

	ws.Route(ws.GET("/{id}").To(u.findUser).
		Doc("get a user").
		Do(tagUsers).
		Param(uid).
		Writes(User{}).
		Returns(200, "OK", User{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.PUT("/{id}").To(u.updateUser).
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

	ws.Route(ws.DELETE("/{id}").To(u.removeUser).
		Doc("delete a user").
		Do(tagUsers).
		Param(uid))

	return ws
}

func (u UserResource) findAllUsers(request *restful.Request, response *restful.Response) {
	list := []User{}
	for _, each := range u.users {
		list = append(list, each)
	}
	response.WriteEntity(list)
}

func (u UserResource) findUser(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	if usr, ok := u.users[id]; !ok {
		response.WriteErrorString(http.StatusNotFound, "User could not be found.")
	} else {
		response.WriteEntity(usr)
	}
}

func (u *UserResource) updateUser(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
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
	id := request.PathParameter("id")
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
	http.Handle(basePath, http.StripPrefix(basePath, http.FileServer(http.Dir("../swagger-ui/dist"))))

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

package handler

import (
	"io"
    //"log"
    "net/http"

    "github.com/emicklei/go-restful"
)


type APIHander struct {
    data Gnocchi
}

func CreateHTTPAPIHandler() (http.Handler, error){
	ids := Ops_paras{"142d8663efce464c89811c63e45bd82e", "123456", 
					"f21a9c86d7114bf99c711f4874d80474", "9f95b9967b894c928880feb32fad1d0d"}
	gnocchi := Gnocchi{ids, ""}
	gnocchi.init()
	apiHandler := APIHander{gnocchi}
	
	wsContainer := restful.NewContainer()
	wsContainer.EnableContentEncoding(true)
	
	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{"X-My-Header"},
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST"},
		CookiesAllowed: false,
		Container:      wsContainer}
	wsContainer.Filter(cors.Filter)
	
    apis := new(restful.WebService)
    apis.
    Path("/monitor_api/v1").
        Consumes( restful.MIME_JSON).
        Produces(restful.MIME_JSON) // you can specify this per route as well

    apis.Route(apis.GET("/topN/{metric}/{num}").To(apiHandler.topN))
	apis.Route(apis.GET("/detail").To(apiHandler.detail))
	apis.Route(apis.GET("/rs_statics").To(apiHandler.rsStatics))
	apis.Route(apis.POST("/set_ops_ids").To(apiHandler.set_ops_ids))

    wsContainer.Add(apis)
	return wsContainer, nil
}


func (u APIHander) topN(request *restful.Request, response *restful.Response) {
    metric := request.PathParameter("metric")
	number := request.PathParameter("num")
	topN_metric := make(map[string]float64)
    ret := u.data.getTopN(metric, number, &topN_metric)
    if (ret != 200) {
        response.AddHeader("Content-Type", "text/plain")
        response.WriteErrorString(http.StatusNotFound, "top5 could not be found.")
    } else {
        response.WriteEntity(topN_metric)
    }
}


func (u APIHander) detail(request *restful.Request, response *restful.Response) {
	 ret, jsonStr := u.data.getDetail()
     if (ret != 200) {
          response.AddHeader("Content-Type", "text/plain")
          response.WriteErrorString(http.StatusNotFound, "detail could not be found.")
      } else {
		 io.WriteString(response, jsonStr)
         //response.WriteEntity((*details).detail_clouds)
      }
}

func (u APIHander) rsStatics(request *restful.Request, response *restful.Response) {
	jsonStr, ret := u.data.getStatics()
	if (ret != 200) {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "detail could not be found.")
	} else {
		io.WriteString(response, jsonStr)
		//response.WriteEntity((*details).detail_clouds)
	}
}

func (u APIHander) set_ops_ids(request *restful.Request, response *restful.Response) {
	ids := new(Ops_paras)
	err := request.ReadEntity(ids)
	if err == nil {
		u.data.set_ops_ids(*ids)
        response.WriteEntity("set sucess")
    } else {
        response.AddHeader("Content-Type", "text/plain")
        response.WriteErrorString(http.StatusInternalServerError, err.Error())
    }
}

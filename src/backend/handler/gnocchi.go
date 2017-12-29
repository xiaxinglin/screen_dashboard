package handler

import (
	"fmt"
    "strconv"
	//"log"
	"net/http"
	"bytes"
	"io/ioutil"

	"github.com/screen_dashboard/src/backend/ini"
	"github.com/screen_dashboard/src/backend/gjson"
)

type Ops_paras struct {
	user_id, password, project_id, domain_id string
}

type Gnocchi struct {
	//topN_metric map[string]usage_metrics
	//details detail_cloud_static
	ids Ops_paras
	token string
}

func get_urls() (string, string, string, error){
	iniPortal, err := iniparser.LoadFile("portal.ini", "utf-8")
	if err != nil {
		//
		return "","","",err
	}
	vip, ok := iniPortal.GetString("openstack", "vip")
	if !ok {
		//
		return "","","",err
	}

	keystone_port, ok := iniPortal.GetString("openstack", "keystone")
	if !ok {
		//
		return "","","",err
	}

	nova_port, ok := iniPortal.GetString("openstack", "nova")
	if !ok {
		//
		return "","","",err
	}

	gnocchi_port, ok := iniPortal.GetString("openstack", "gnocchi")
	if !ok {
		//
		return "","","",err
	}
	url_auth := fmt.Sprintf("http://%s:%d/v3/auth/tokens", vip, keystone_port)
	url_nova := fmt.Sprintf("http://%s:%d", vip, nova_port)
	url_gnocchi := fmt.Sprintf("http://%s:%d", vip, gnocchi_port)
	return url_auth, url_nova, url_gnocchi, err
}

func  get_token(ids Ops_paras) (string, error){
	s := fmt.Sprintf(`{"auth": {"identity": {"methods": ["password"],"password": {
        "user": {"id":"%s", "password": "%s"}}},
        "scope":{"project": {"domain": {"name": "%s"},"id": "%s"}}} 
        }`,ids.user_id, ids.password, ids.domain_id, ids.project_id)

	data := []byte(s)
	body := bytes.NewReader(data)
	authUrl,_,_,err := get_urls()
	if err != nil {
		return "", err
	}

    resp, err := http.Post(authUrl,"application/json;charset=utf-8", body)
    if err != nil {
        return "", err
    }

    out := resp.Header.Get("X-Subject-Token")
    return out, err
}

func (gno Gnocchi) init() {
	gno.token, _ = get_token(gno.ids)
}

func (gno Gnocchi) getTopN(metric string, number string, out *map[string]float64) int {
	if gno.ids.user_id == "" {
		return -1
	}
	num, err := strconv.Atoi(number)
	if err != nil {
		panic(err)
		return -1
	} 

	ret :=  gno.getGnocchiTopN(metric, num, out)

	if ret == 401 {
		gno.token, _ = get_token(gno.ids)
		ret = gno.getGnocchiTopN(metric, num, out)
	}

	return ret
}

func (gno Gnocchi)  getGnocchiTopN(metric string, num int, out  *map[string]float64) int {
	_,_,gnoUrl,err := get_urls()
	if err != nil {
		return  -1
	}

	gnoTopNUrl := fmt.Sprintf("%s/v1/topn?num=%d&metric=%s", gnoUrl, num, metric)
	client := &http.Client{}
	request, err := http.NewRequest("GET", gnoTopNUrl, nil)
	request.Header.Set("X-Auth-Token", gno.token)
	response, _ := client.Do(request)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// handle error
	}
	status := response.StatusCode
	fmt.Println(string(body))
	jsonStr := string(body)

	name := gjson.Get(jsonStr, `topn`)
	lenTopn := len(name.Array())
	i:= 0
	for i < lenTopn {
		indexStrValue := fmt.Sprintf("topn.%d.resource_data.metric_value", i)
		indexStrHost := fmt.Sprintf("topn.%d.resource_data.name", i)
		hostName := gjson.Get(jsonStr, indexStrHost).String()
		(*out)[hostName] = gjson.Get(jsonStr, indexStrValue).Float()
		i = i + 1
	}

	return status
}

func (gno Gnocchi) getHostStatics()  (string, int, error) {
	_,novaUrl,_,err := get_urls()
	if err != nil {
		return "", -1, err
	}

	novaUrl = novaUrl + "/v2.1/" + gno.ids.project_id + "/os-hypervisors/statistics"
	client := &http.Client{}
	request, err := http.NewRequest("GET", novaUrl, nil)
	request.Header.Set("X-Auth-Token", gno.token)
	response, _ := client.Do(request)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// handle error
	}
	status := response.StatusCode
	fmt.Println(string(body))

	return string(body), status, err
}

func (gno Gnocchi) getStatics() (string, int) {
	if gno.ids.user_id == "" {
		return "", -1
	}

	jsonStr, ret, err :=  gno.getHostStatics()

	if err != nil {
		return "", -1
	}

	if ret == 401 {
		gno.token, _ = get_token(gno.ids)
		jsonStr, ret, err = gno.getHostStatics()
	}

	return jsonStr, ret
}

func (gno Gnocchi) getDetail() (int, string) {
	if gno.ids.user_id == "" {
        return -1, ""
    }

	ret, jsonStr :=  gno.getHostDetail()

    if ret == 401 {
        gno.token, _ = get_token(gno.ids)
        ret, jsonStr = gno.getHostDetail()
    }	

	return ret, jsonStr
}

func (gno Gnocchi)  getHostDetail() (int, string)  {
    /*detail := new(Detail_cloud)
    detail.servername = "hahah"
	detail.cloud_score = 34
    (*out).detail_clouds["hehe"] = *detail*/
	_,_,gnoUrl,err := get_urls()
	if err != nil {
		return  -1, ""
	}

	gnoTopNUrl := fmt.Sprintf("%s/v1/all_measures", gnoUrl)
	client := &http.Client{}
	request, err := http.NewRequest("GET", gnoTopNUrl, nil)
	request.Header.Set("X-Auth-Token", gno.token)
	response, _ := client.Do(request)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// handle error
	}
	status := response.StatusCode
	fmt.Println(string(body))
	jsonStr := string(body)

	jsonState := ""
	num_excellent := 0
	num_good := 0
	num_poor := 0
	var sum int
	name := gjson.Get(jsonStr, `measures`)
	sum = len(name.Array())
	i := 0
	for i < sum {
		hostStr := fmt.Sprintf("measures.%d.resource_data.measures.name", i)
		memUsedStr := fmt.Sprintf("measures.%d.resource_data.measures.hardware.memory.used", i)
		memTotalStr := fmt.Sprintf("measures.%d.resource_data.measures.hardware.memory.total", i)
		cpuUtilStr := fmt.Sprintf("measures.%d.resource_data.measures.hardware.cpu.util", i)
		loadStr := fmt.Sprintf("measures.%d.resource_data.measures.hardware.cpu.load.5min", i)

		hostname := gjson.Get(jsonStr, hostStr).Int()
		memUsed := gjson.Get(jsonStr, memUsedStr).Int()
		memTotal := gjson.Get(jsonStr, memTotalStr).Int()
		cpuUtil := gjson.Get(jsonStr, cpuUtilStr).Float()
		load := gjson.Get(jsonStr, loadStr).Float()

		scroe := float64(100) - float64(memUsed/memTotal)*10 - cpuUtil*5 - load/100*5
		if scroe >= 90 {
			num_excellent = num_excellent + 1
		} else if scroe >=70 {
			num_good = num_good + 1
		} else {
			num_poor = num_poor + 1
		}

		jsonTemp :=  fmt.Sprintf(`{"%s":%f}`, hostname, scroe)
		if jsonState != "" {
			jsonState = jsonState + ",\n" + jsonTemp
		}else{
			jsonState = jsonTemp
		}

		i = i + 1
	}

	system_score :=  float64(100) - float64(num_excellent/sum)*3 -float64(num_good/sum)*5 - float64(num_poor/sum)*10
	jsonOutput := fmt.Sprintf(`"hostState":{"sum":%d, "system_score" : %d, "num_excellent" :%d,
					"num_good":%d, "num_poor":%d, "details":[%s]}`, sum, int(system_score), int(num_excellent),
					int(num_good), int(num_poor), jsonState)

	return  status, jsonOutput
}

func (gno Gnocchi)  set_ops_ids(ids Ops_paras) {
	gno.ids.project_id = ids.project_id
	gno.ids.password = ids.password
	gno.ids.user_id = ids.user_id
	gno.ids.domain_id = ids.domain_id
}

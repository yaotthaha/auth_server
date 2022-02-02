package main

import (
	"auth_server/lib"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func HTTPServer(ListenAddr string, ListenPort int) {
	/**
	Request Param:
	@ ListenAddr string
	@ ListenPort uint16
	Respon Param:
	StatusCode:
	(-26) HTTP Server Start Fail
	*/
	http.HandleFunc("/", route)
	err := http.ListenAndServe(ListenAddr+":"+strconv.Itoa(ListenPort), nil)
	if err != nil {
		OutputLog(-1, "HTTP Server Start Fail: "+err.Error())
	}
}

func IP46Standard(IP string) string {
	if strings.Contains(IP, ":") {
		return "[" + IP + "]"
	} else {
		return IP
	}
}

func ClientIPGet(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return IP46Standard(ip)
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return IP46Standard(ip)
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return IP46Standard(ip)
	}
	ipArray := strings.Split(r.RemoteAddr, ":")
	if len(ipArray) <= 2 {
		// IPv4
		return IP46Standard(ip)
	} else {
		// IPv6
		ip = ""
		for i := 0; i < len(ipArray)-1; i++ {
			ip = ip + ipArray[i] + ":"
		}
		ip = ip[0 : len(ip)-1]
		return IP46Standard(ip)
	}
}

func route(w http.ResponseWriter, r *http.Request) {
	Method, ArgsALL := HttpMethodParams(r)
	RequestInfo := make(map[string]interface{})
	RequestInfo["ip"] = ClientIPGet(r)
	RequestInfo["url"] = strings.Split(r.RequestURI, "?")[0]
	RequestInfo["ua"] = r.UserAgent()
	RequestInfo["method"] = Method
	RequestInfo["params"] = ArgsALL
	RequestInfoJSON, _ := json.Marshal(RequestInfo)
	OutputLog(2, "[HTTP Server] New Request: "+string(RequestInfoJSON))
	// BlackList Code
	switch Method {
	case "GET", "POST":
	case "OPTIONS":
		w.Header().Set("Access-Control-Max-Age", "7200")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		return
	default:
		OutputLog(1, "[HTTP Server] Invalid Request Method: "+Method)
		_, _ = w.Write([]byte("Invalid Request"))
		w.WriteHeader(404)
		return
	}
	switch RequestInfo["url"] {
	case "/auth":
		if !Auth_Open_Status {
			_, _ = w.Write([]byte(HTTPJSONRespon("999", "Server Close")))
			w.WriteHeader(403)
			return
		}
		HTTPAuthResponFunc(w, RequestInfo)
		return
	case "/admin":
		if !Admin_Open_Status {
			_, _ = w.Write([]byte(HTTPJSONRespon("999", "Server Close")))
			w.WriteHeader(403)
			return
		}
		if !AdminAccess(RequestInfo["ip"].(string)) {
			_, _ = w.Write([]byte("403"))
			w.WriteHeader(403)
			return
		}
		HTTPAdminResponFunc(w, RequestInfo)
		return
	case "/timestamp":
		_, _ = w.Write([]byte(func() string {
			return lib.GetTimestamp_S_String()
		}()))
		return
	default:
		OutputLog(1, "[HTTP Server] Invalid Request: "+RequestInfo["url"].(string))
		_, _ = w.Write([]byte("Invalid Request"))
		w.WriteHeader(404)
		return
	}
}

func HttpMethodParams(r *http.Request) (string, map[string]string) {
	ArgsALL := make(map[string]string)
	arg := r.URL.Query()
	for k, v := range arg {
		ArgsALL[k] = v[len(v)-1]
	}
	DataPost, _ := ioutil.ReadAll(r.Body)
	DataPost_Map := make(map[string]string)
	_ = json.Unmarshal(DataPost, &DataPost_Map)
	for k, v := range DataPost_Map {
		ArgsALL[k] = v
	}
	return r.Method, ArgsALL
}

func HTTPAuthResponFunc(w http.ResponseWriter, RequestInfo map[string]interface{}) {
	ConnectionTag := lib.GenRandomString(12) + "_#_" + RequestInfo["ip"].(string)
	ArgsJSONString, _ := json.Marshal(RequestInfo["params"])
	OutputLog(0, lib.JoinString("[HTTP Server] New Auth Request [", ConnectionTag, "] IP: ", RequestInfo["ip"].(string), " {", RequestInfo["method"].(string), "} ", RequestInfo["url"].(string), " ", string(ArgsJSONString)))
	AuthRespon := AuthRoute(RequestInfo["params"].(map[string]string), ConnectionTag)
	OutputLog(0, lib.JoinString("[HTTP Server] Auth [", ConnectionTag, "] Respon: ", AuthRespon))
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(AuthRespon))
	return
}

func HTTPAdminResponFunc(w http.ResponseWriter, RequestInfo map[string]interface{}) {
	ConnectionTag := lib.GenRandomString(12) + "_#_" + RequestInfo["ip"].(string)
	ArgsJSONString, _ := json.Marshal(RequestInfo["params"])
	OutputLog(0, lib.JoinString("[HTTP Server] New Admin Request: [", ConnectionTag, "] IP: ", RequestInfo["ip"].(string), " {", RequestInfo["method"].(string), "} ", RequestInfo["url"].(string), " ", string(ArgsJSONString)))
	AdminRespon := AdminRoute(RequestInfo["params"].(map[string]string), ConnectionTag)
	OutputLog(0, lib.JoinString("[HTTP Server] Admin [", ConnectionTag, "] Respon: ", AdminRespon))
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(AdminRespon))
	return
}

func AdminAccess(IP string) bool {
	if !Admin_CIDR_Access_Status {
		return true
	}
	return AccessIP(Admin_Access_CIDR_Set, IP)
}

func AccessIP(IPArray []string, IP string) bool {
	if strings.Contains(IP, ":") {
		IP = IP[1 : len(IP)-1]
	}
	ip := net.ParseIP(IP)
	for _, v := range IPArray {
		_, ipnet, _ := net.ParseCIDR(v)
		if ipnet.Contains(ip) {
			return true
		}
	}
	return false
}

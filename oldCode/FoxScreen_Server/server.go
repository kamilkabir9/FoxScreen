package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var VideoDir = "CropResult"

//Mac is Base64 of Mac address
type Mac string //TODO Mac must be replaced by uuid(handshake gives device with uuid)

type Knocks struct {
	Loc   string
	Epoch float64 //TODO Epoch is do slow.need faster Epoch
	Mac
}

var AKMutex = &sync.Mutex{}
var AllKnocks = make([]Knocks, 0)
var llog = log.New(os.Stdout, ">", log.Lshortfile)

func addKnocks(k Knocks) {
	AKMutex.Lock()
	defer AKMutex.Unlock()
	AllKnocks = append(AllKnocks, k)
}

//Find_MatchKnks gooes thru all stored knocks at returns knocks that have same knocks.Epoch
func Find_MatchKnks() ([]Knocks, bool) {
	fmt.Println("stat Find_Match")
	defer fmt.Println("END Find_Match")
	AKMutex.Lock()
	defer AKMutex.Unlock()
	rsltKnk := make([]Knocks, 0)
	rsltfound := false
	for _, k1 := range AllKnocks {
		for _, k2 := range AllKnocks {
			fmt.Printf("%v==?%v\n", k1, k2)
			if (k1.Epoch == k2.Epoch) && (k1 != k2) {
				llog.Println("Yes")
				rsltfound = true
				rsltKnk = append(rsltKnk, k1, k2)
				break
			}
		}
		if rsltfound == true {
			break
		}
	}
	fmt.Println("rsltKnk: ", rsltKnk)
	llog.Println("Before", AllKnocks)
	if rsltfound {
		temp := make([]Knocks, 0)
		for _, k := range AllKnocks {
			if k != rsltKnk[0] && k != rsltKnk[1] {
				temp = append(temp, k)
				fmt.Println(temp)
			}
		}
		AllKnocks = temp
	}
	llog.Println("After", AllKnocks)
	return rsltKnk, rsltfound
}

func ListKnocks() (string, error) {
	AKMutex.Lock()
	defer AKMutex.Unlock()
	if len(AllKnocks) == 0 {
		llog.Println("Zero Knocks")
		return "", errors.New("Zero Knocks")
	}
	//fmt.Println(RcvdKnocks)
	result, err := json.Marshal(AllKnocks)
	return string(result), err
}

type Neighbours struct {
	N []Device
	S []Device
	W []Device
	E []Device
}

type Device struct {
	Height, Width float64
	Orientation   int
	Neighbours
	Mac
	epoch int
}

func (d *Device) printinfo() {
	fmt.Println(d)
}

var ADMutex = &sync.Mutex{}
var AllDevice = make(map[Mac]*Device, 1)

func addDevice(dev *Device) error {
	ADMutex.Lock()
	defer ADMutex.Unlock()
	_, found := AllDevice[dev.Mac]
	if found {
		return errors.New(fmt.Sprintf("Device %v already registerd", dev.Mac))
	}
	AllDevice[dev.Mac] = dev
	return nil
}

func ListDevice() (string, error) {
	ADMutex.Lock()
	defer ADMutex.Unlock()
	if len(AllDevice) == 0 {
		llog.Println("Zero Device")
		return "", errors.New("Zero Device")
	}
	//fmt.Println(AllDevice)
	result, err := json.Marshal(AllDevice)
	return string(result), err
}
func pairDevice(k1 Knocks, k2 Knocks) error {
	ADMutex.Lock()
	defer ADMutex.Unlock()
	//	TODO below command returns a copy but we need a refernse to tht struct so make map[mac]*device

	_, found := AllDevice[k1.Mac]
	if !found {
		return errors.New("matching Device not found for this knock k1")
	}
	_, found = AllDevice[k2.Mac]
	if !found {
		return errors.New("matching Device not found for this knock k2")
	}
	return nil
}

var upgrader = websocket.Upgrader{} // use default options

//TODO R2Json is refactord change in App
//R2Json takes to two string and JSON's it
//its useful when returning successes or error to REST endpoints
func R2Json(ttype string, msg string) string {
	j := make(map[string]string)
	j["TType"] = ttype
	j["Msg"] = msg
	jjson, _ := json.Marshal(j)
	return string(jjson)
}

// {"Mac":"YTA6MGI6YmE6Mjc6ZmM6MWM=","Width":"720","Height":"1184","Orientation":"1"}
type HandshakeJson struct {
	Mac
	Width       string
	Height      string
	Orientation string
}

func (h *HandshakeJson) AsDevice() (Device, error) {
	var err error
	device := Device{}
	device.Height, err = strconv.ParseFloat(h.Height, 32)
	if err != nil {
		return device, err
	}
	device.Width, err = strconv.ParseFloat(h.Width, 32)
	if err != nil {
		return device, err
	}
	device.Mac = h.Mac
	device.Orientation, err = strconv.Atoi(h.Orientation)
	if err != nil {
		return device, err
	}
	device.epoch = int(time.Now().Unix())
	return device, nil
}
func HandShake_Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HandShake_Handler START")
	defer fmt.Println("HandShake_Handler END")
	val := r.URL.Query()
	if len(val) <= 0 {
		fmt.Fprint(w, R2Json("err", "No Values given to handShake"))
	} else {
		dev := Device{}
		// {"Mac":"YTA6MGI6YmE6Mjc6ZmM6MWM=","Width":"720","Height":"1184","Orientation":"1"}
		dev.Mac = Mac(val.Get("Mac"))
		dev.Width, _ = strconv.ParseFloat(val.Get("Width"), 32)
		dev.Height, _ = strconv.ParseFloat(val.Get("Height"), 32)
		dev.Orientation, _ = strconv.Atoi(val.Get("Orientation"))
		err := addDevice(&dev)
		if err != nil {
			fmt.Fprint(w, R2Json("err", fmt.Sprint(err)))
		} else {
			fmt.Fprint(w, R2Json("true", "Added device to Alldevice"))
		}
	}
}

//returns TType=err OR true
func Knocks_Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Knocks_Handler START")
	defer fmt.Println("Knocks_Handler END")
	val := r.URL.Query()
	if len(val) <= 0 {
		fmt.Fprint(w, R2Json("err", "No Values given to Knock"))
	} else {
		Newknk := Knocks{}
		Newknk.Mac = Mac(val.Get("Mac"))
		Newknk.Loc = val.Get("Loc")
		Newknk.Epoch, _ = strconv.ParseFloat(val.Get("Epoch"), 32)
		addKnocks(Newknk)
		fmt.Fprint(w, R2Json("true", "Added Knock to Knock"))
		knks, found := Find_MatchKnks()
		if !found {
			llog.Println("didnt find any mathcing knk")
			//	TODO send meaningfull message to device tht no matching knock
		}
		llog.Println("Found these knks:", knks)

		//a, err := ListDevice()
		//if err != nil {
		//	llog.Println("err", err)
		//}
		//fmt.Println(a)

	}

}

/*
MStream_Handler Sends
1-Err
2-Download link
3-Epoch time
*/
func MStream_Handler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		llog.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		var dat map[string]string

		if err := json.Unmarshal(message, &dat); err != nil {
			panic(err)
		}

		if err != nil {
			llog.Println("read:", err)
			break
		}
		fmt.Printf("recv: %s", dat)
		err = c.WriteMessage(mt, message)
		if err != nil {
			fmt.Println("write:", err)
			break
		}
	}
}

//serve requested .mp4 file
//localhost:8080/DownloadMp4?FileID=xxxx.mp4
func DownloadMp4_Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DownloadMp4_Handler START")
	defer fmt.Println("DownloadMp4_Handler END")
	val := r.URL.Query()
	FileID := val.Get("FileID")
	llog.Println("requested mp4: ", FileID)
	if FileID == "" {
		fmt.Fprint(w, R2Json("err", "No FileID"))
		return
	}
	FileID = strings.Join([]string{VideoDir, FileID}, string(os.PathSeparator))
	http.ServeFile(w, r, FileID)
}

func WebView_DeviceList_Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebView_DeviceList_Handler START")
	defer fmt.Println("WebView_DeviceList_Handler END")
	deviceList, err := ListDevice()
	if err != nil {
		fmt.Fprint(w, R2Json("err", err.Error()))
		return
	}
	fmt.Fprint(w, R2Json("true", string(deviceList)))
}

func WebView_KnocksList_Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebView_KnocksList_Handler START")
	defer fmt.Println("WebView_KnocksList_Handler END")
	deviceList, err := ListKnocks()
	if err != nil {
		fmt.Fprint(w, R2Json("err", err.Error()))
		return
	}
	fmt.Fprint(w, R2Json("true", string(deviceList)))
}

/*Gets Deploy command From FrontEnd.This starts
1-cropping Video according to devices
2-Sends download links to all device
3-Waits for all Devices to replay about finished downloads
4-sends epoch time to start playing the video in all devices
*/
func WebView_Deploy_Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebView_Deploy_Handler START")
	defer fmt.Println("WebView_Deploy_Handler END")
	deviceList, err := ListKnocks()
	if err != nil {
		fmt.Fprint(w, R2Json("err", err.Error()))
		return
	}
	fmt.Fprint(w, R2Json("true", string(deviceList)))
}

func main() {
	llog.Println("Running FoxScreen Server")
	/*URL points for Websockets
	1- /HandShake Replay from this handShake starts other sockets(/knocks,MStream)
	2- /Knocks receive Knocks from devices AND sends back boolean if a neighbour is found
	3- /MStream sents .mp4 or H264 video file
	4- /WebView sends Data for visualization and Receive commands to execute .
	*/

	http.HandleFunc("/HandShake", HandShake_Handler)
	http.HandleFunc("/Knock", Knocks_Handler)

	http.HandleFunc("/MStream", MStream_Handler) //WebSocket
	http.HandleFunc("/DownloadMp4", DownloadMp4_Handler)

	http.HandleFunc("/WebView/DeviceList", WebView_DeviceList_Handler)
	http.HandleFunc("/WebView/KnocksList", WebView_KnocksList_Handler)
	http.HandleFunc("/WebView/Deploy", WebView_Deploy_Handler)

	http.Handle("/", http.FileServer(http.Dir("./")))

	llog.Fatal(http.ListenAndServe(":8080", nil))
}

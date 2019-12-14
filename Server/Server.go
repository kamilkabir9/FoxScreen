//args[1]=mediaType args[2]=mediaFileLoc
//Example go run -race Server.go pic Media/flag.png
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)
var gmLog log.Logger

//webSocket message
type MshipConnectionWSmsg struct {
	JobType  string
	Data     string
	DeviceId int `json:"deviceId,string"`
}

func (m MshipConnectionWSmsg) json() string {
	b, err := json.Marshal(m)
	if err != nil {
		log.Println("err", err)
		errJson, _ := json.Marshal(MshipConnectionWSmsg{err.Error(), "", m.DeviceId})
		return string(errJson)
	}
	return string(b)
}

func makeMSWSmsg(msg []byte) MshipConnectionWSmsg {
	var result MshipConnectionWSmsg
	err := json.Unmarshal(msg, &result)
	if err != nil {
		log.Println("Err", err)
	}
	return result
}

////returns Media File according to DeviceID
//func getMedia(deviceId string) string {
//	//TODO get corresponding video of deviceid
//
//	return "flag.mp4"
//}

//device counter
var DC int = 0
var DCMux sync.Mutex

type Device struct {
	ID    int
	Width int
	//summation of neighbor(s) width
	PrevWidth int
	//summation of neighbor(s) Height
	PrevHeight      *int
	Height          int
	North_Neighbour *Device
	South_Neighbour *Device
	West_Neighbour  *Device
	East_Neighbour  *Device
}

func (d *Device) AddNighbour(loc string, neighbor_Device *Device) {
	switch loc {
	case "North":
		d.North_Neighbour = neighbor_Device
	case "South":
		d.South_Neighbour = neighbor_Device
	case "West":
		d.West_Neighbour = neighbor_Device
	case "East":
		d.East_Neighbour = neighbor_Device
	default:
		log.Println("ERR-->Unknown loc:", loc)
	}
}
func getIdOfDevice(d *Device) string {
	if d != nil {
		return fmt.Sprint(d.ID)
	}
	return "nil"
}
func (d Device) String() string {
	return fmt.Sprintf("-------\nID:%v\tWidth:%v\tHeight:%v\nNorth:%v\tSouth:%v\nWest:%v\tEast:%v\nPrevWidth:%v\tPrevHeight:%v\n-------\n", d.ID, d.Width, d.Height, getIdOfDevice(d.North_Neighbour), getIdOfDevice(d.South_Neighbour), getIdOfDevice(d.West_Neighbour), getIdOfDevice(d.East_Neighbour), d.PrevWidth, *d.PrevHeight)
	//	TODO neighbours
}

type RowofDevice struct {
	//prevRow *RowofDevice
	devices []*Device
	width   int
	height  int
	//abscissa int
	ordinate int
}

var TableOfDevice = make(map[int]RowofDevice)

//All the device connected stored as map[int]*Device
var conctdDevices sync.Map //map[int]*Device

func printConctdDevices() {
	//for d:=range conctdDevices.Range()
	fmt.Println("-------------------")
	conctdDevices.Range(func(deviceID, device interface{}) bool {
		fmt.Println("deviceID:", deviceID)
		fmt.Println(device)
		fmt.Println("-------------------")
		return true
	})
}

//var WSMux sync.Mutex
//var  int = 0
type cropResultChnlStruct struct {
	MediaType  string
	FileName   string
	receiverID int
}

//TODO increase buffer ?
var cropResultChanList sync.Map

//------knock-------//

//goodDurKnock : if two knocks are recived within goodDurKnock then knocks are matched
var goodDurKnock = time.Duration(time.Second * 3)

// ConfirmedKnockPairCount: count the Total number confirmed knock pair count
var ConfirmedKnockPairCount int = 0
var CKPCMux sync.Mutex

// knock
type knock struct {
	Id  int    `json:"id,string"`
	Loc string `json:"loc"`
	//JobType = knock or connect
	//JobType  string `json:"jobType,string"`
	JobType  string    `json:"jobType"`
	TimeSend time.Time `json:"timeSend,string"`
	TimeRcvd time.Time `json:"timeRcvd,string"`
}

func (k knock) String() string {
	return fmt.Sprintf("---------------\nId:%v,Loc:%v,JobType:%v,\nTimeSend:%v,TimeRcvd:%v\n-----------\n", k.Id, k.Loc, k.JobType, k.TimeSend, k.TimeSend)
}

func makeKWSmsg(msg []byte) knock {
	var result knock
	err := json.Unmarshal(msg, &result)
	if err != nil {
		log.Println("Err", err)
	}
	return result
}

// confirmedKnock : k1's id(or device) are k2's id(or device) is considered neighbours if delay <
type confirmedKnock struct {
	Id        int    `json:"id,string"`
	Loc       string `json:"loc"`
	ConfirmId int
}

//confirmedKnockChanList : chan used to send all confirmed knock pairs
var confirmedKnockChanList sync.Map

//recvdKnock : collection of unpaired knocks .matched knocks will be removed
var recvdKnocks = struct {
	Knocks  map[int]knock
	MX      sync.Mutex
	Counter int
}{Knocks: make(map[int]knock)}

//------knock-------//

var mediaType, mediaFile string

func main() {
	log.SetFlags(log.Lshortfile)
	gmLogFile, err := os.Create("gmLog.txt")
	mediaType = os.Args[1]
	if mediaType != "pic" && mediaType != "video" {
		log.Fatal("Wrong media type given as input")
	}
	mediaFile = os.Args[2]
	if err != nil {
		log.Fatal(err)
	}
	gmLog.SetOutput(gmLogFile)
	fmt.Println("Starting FoxScreen")
	http.HandleFunc("/GetDeviceID", CreateNewDeviceIDHandler)
	http.HandleFunc("/MshipWSConnection", MshipConnectionWSHandler)
	http.HandleFunc("/KnockWSConnection", KnockWSHandler)
	http.HandleFunc("/display", displayHandler)
	http.Handle("/", http.FileServer(http.Dir("../Frontend")))
	log.Fatal(http.ListenAndServe(":8080", nil))
	//log.Fatal(http.ListenAndServe(":8080",http.Handle("/", http.FileServer(http.Dir("/tmp")))))
}
func displayHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../Frontend/display.html")
}
func MshipConnectionWSHandler(w http.ResponseWriter, r *http.Request) {
	var thisUniqID int
	//fmt.Println(thisUniqID)
	fmt.Println("starting MshipConnectionWS conn")
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		fmt.Println("msg:", string(msg))
		msgRcvd := makeMSWSmsg(msg)
		switch msgRcvd.JobType {
		case "connect":
			thisUniqID = msgRcvd.DeviceId
			err := conn.WriteJSON(MshipConnectionWSmsg{"connected", "", thisUniqID})
			if err != nil {
				log.Println(err)
				break
			}
			cropResultChanList.Store(thisUniqID, make(chan cropResultChnlStruct))
			go func(conn *websocket.Conn, id int) {
				cropResultChanG, found := cropResultChanList.Load(id)
				if !found {
					log.Println("ERR--> cropResultChanList couldnt find chan with id:", id)
					return
				}
				cropResultChan := cropResultChanG.(chan cropResultChnlStruct)
				cropRslt := <-cropResultChan
				//Not good way to do it 1.func exits after recving 1st wrong packet in chan 2.sending back the received RESULT might loop between the same sender and receiver
				//if id != cropRslt.receiverID {
				//	log.Println("id!=cropRslt.receiverID")
				//	//	sending back the wrong addressed msg back to channel to be recived at correct address
				//	cropResultChanList <- cropRslt
				//	return
				//}
				fmt.Println("sending ", cropRslt, "to", id)
				err := conn.WriteJSON(MshipConnectionWSmsg{"crop_result_" + cropRslt.MediaType, cropRslt.FileName, id})
				if err != nil {
					log.Println(err)
				}

			}(conn, thisUniqID)

		case "crop_start":
			//err = conn.WriteMessage(msgType, []byte("crop msg received.Starting cropping"))
			err := conn.WriteJSON(MshipConnectionWSmsg{msgRcvd.JobType, "working", thisUniqID})
			fmt.Println("crop msg received.Starting cropping")
			if err != nil {
				log.Println(err)
				break
			}
			cropMachine()
		default:
			err := conn.WriteJSON(MshipConnectionWSmsg{"err", "Wrong Job Type !!!", thisUniqID})
			if err != nil {
				log.Println(err)
				break
			}
		}
	}
	conn.Close()
	//return
}

func KnockWSHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("starting KnockWS conn")
	var thisUniqID int
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		fmt.Println("msg:", string(msg))
		msgRcvd := makeKWSmsg(msg)
		switch msgRcvd.JobType {
		case "connect":
			thisUniqID = msgRcvd.Id
			confirmedKnockChanList.Store(thisUniqID, make(chan confirmedKnock, 50))
			fmt.Println("Knock websocket connection made to Device : ", thisUniqID)
			//TODO listen on chan for matched knocks
			err := conn.WriteJSON(MshipConnectionWSmsg{"connected", "", thisUniqID})
			if err != nil {
				log.Println(err)
				break
			}
			go func(conn *websocket.Conn, id int) {
				knockedChanG, found := confirmedKnockChanList.Load(thisUniqID)
				if !found {
					log.Println("knockedChan chan not found for ID", thisUniqID)
				}
				knockedChan := knockedChanG.(chan confirmedKnock)
				for {
					knock := <-knockedChan
					if id == knock.Id {
						fmt.Println("sending ", knock, "to", id)
						err := conn.WriteJSON(knock)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}(conn, thisUniqID)

		case "knock":

			recvdKnocks.MX.Lock()
			recvdKnocks.Counter += 1
			msgRcvd.TimeRcvd = time.Now()
			recvdKnocks.Knocks[recvdKnocks.Counter] = msgRcvd
			fmt.Println("current Knocks")
			for _, k := range recvdKnocks.Knocks {
				fmt.Print(k)
			}
			FindMatchingKnocks()
			recvdKnocks.MX.Unlock()
			//err := conn.WriteJSON(MshipConnectionWSmsg{msgRcvd.JobType, "working", thisUniqID})
			//fmt.Println("crop msg received.Starting cropping")
			//if err != nil {
			//	log.Println(err)
			//	break
			//}
		default:
			err := conn.WriteJSON(MshipConnectionWSmsg{"err", "Wrong Job Type !!!", thisUniqID})
			if err != nil {
				log.Println(err)
				break
			}
		}
	}
	conn.Close()
	//return
}
func CreateNewDeviceIDHandler(w http.ResponseWriter, r *http.Request) {
	DCMux.Lock()
	defer DCMux.Unlock()
	DC = DC + 1
	fmt.Fprint(w, DC)
	log.Println("added Device : ", DC)
	//TODO redirect to diff log file
	deviceWidth, err := strconv.Atoi(r.URL.Query().Get("width"))
	if err != nil {
		log.Println("err", err)
	}
	deviceHeight, err := strconv.Atoi(r.URL.Query().Get("height"))
	if err != nil {
		log.Println("err", err)
	}
	//conctdDevices.Store(DC, &Device{DC, deviceWidth, 0,new(int),deviceHeight, new(Device), new(Device), new(Device), new(Device)})
	conctdDevices.Store(DC, &Device{DC, deviceWidth, 0, new(int), deviceHeight, nil, nil, nil, nil})
	printConctdDevices()
}

func cropMachine() {
	fmt.Println("starting croping")
	startTime := time.Now()
	//time.Sleep(time.Second * 2)
	cropWizard()
	fmt.Println("Cropping took", time.Since(startTime).String())

	DCMux.Lock()
	defer DCMux.Unlock()
	for i := 1; i <= DC; i++ { //TODO use conctdDevices.Range(func(deviceID,device interface{})bool{}) ???
		fmt.Println("sent crop to device : ", i)
		fileName := path.Join(path.Dir(mediaFile), strings.Replace(path.Base(mediaFile), path.Ext(mediaFile), "-"+strconv.Itoa(i)+path.Ext(mediaFile), 1))
		cropResultChanG, found := cropResultChanList.Load(i)
		if found {
			cropResultChan := cropResultChanG.(chan cropResultChnlStruct)
			cropResultChan <- cropResultChnlStruct{mediaType, fileName, i}
		} else {
			log.Println("ERR--> cropResultChanG not found for id ", i)
		}
		//cropResultChanList <- cropResultChnlStruct{"video","flag.mp4"}
	}
}

func cropWizard() {

	//creating table
	//get the Col_device at left-top  corner
	deviceG, found := conctdDevices.Load(1)
	if !found {
		log.Println("Err-->Col_device not found with id :", 1)
	}
	Col_device := deviceG.(*Device)
	row := 0
	var ordinate = 0
	for Col_device != nil {
		fmt.Println("Row -->", row)
		row_Width := 0
		var row_Height = new(int)
		east_neighbor := Col_device
		var east_nieghbours_Row []*Device
		for east_neighbor != nil {
			fmt.Println("east id->", east_neighbor.ID)
			east_neighbor.PrevWidth = row_Width
			east_neighbor.PrevHeight = row_Height
			row_Width += east_neighbor.Width
			if *row_Height < east_neighbor.Height {
				*row_Height = east_neighbor.Height
			}
			east_nieghbours_Row = append(east_nieghbours_Row, east_neighbor)
			east_neighbor = east_neighbor.East_Neighbour
		}
		ordinate += *row_Height
		thisRow := RowofDevice{east_nieghbours_Row, row_Width, *row_Height, int(math.Abs(float64(ordinate - *row_Height)))}
		TableOfDevice[row] = thisRow
		row += 1
		Col_device = Col_device.South_Neighbour
	}

	fmt.Println("===============TableOfDevice===============")
	for i, v := range TableOfDevice {
		fmt.Println("=============")
		fmt.Println("Row :", i)
		fmt.Printf("Width:%v\t Height:%v\t ordinate:%v\n ", v.width, v.height, v.ordinate)
		fmt.Println("DEvices-->")
		for _, d := range v.devices {
			fmt.Println(d)
		}
		fmt.Println("=============")
	}
	table_width, table_height := getTableData(TableOfDevice)
	//Cropping the Pic
	//Type 1
	resizedFileName := strings.Replace(path.Base(mediaFile), path.Ext(mediaFile), "-Resized"+path.Ext(mediaFile), 1)
	//TODO handle video use ffmpeg
	resizeMediaCommnd := exec.Command("convert", path.Base(mediaFile), "-verbose", "-resize", fmt.Sprintf("%vx%v!", table_width, table_height), resizedFileName)
	//Type 2
	//cpCmd := exec.Command("cp", path.Base(mediaFile), strings.Replace(path.Base(mediaFile), path.Ext(mediaFile), "-Original"+path.Ext(mediaFile), 1))
	//cpCmd.Dir = "../Frontend/Media"
	//err := cpCmd.Run()
	//if err != nil {
	//	log.Println("ERR:",err)
	//}
	//resizedFileName := path.Base(mediaFile)
	//mogrify -resize 1000x2200 -background white -gravity northwest -extent 2000x1200 testImage.png
	//http://cubiq.org/create-fixed-size-thumbnails-with-imagemagick
	//resizeMediaCommnd := exec.Command("mogrify",  "-verbose", "-resize", fmt.Sprintf("%vx%v", table_width, table_height),"-background"," white","-gravity","northwest","-extent",fmt.Sprintf("%vx%v", table_width, table_height), resizedFileName)
	resizeMediaCommnd.Dir = "../Frontend/Media"
	fmt.Println(resizeMediaCommnd.Args)
	output, err := resizeMediaCommnd.CombinedOutput()
	if err != nil {
		log.Println("err", err)
	}
	fmt.Println(string(output))

	nThCropedFileName := ""
	//TODO needed ?
	DCMux.Lock()
	defer DCMux.Unlock()
	//for i := 1; i <= DC; i++ {
	for _, this_Row := range TableOfDevice {
		//this_Row_width:=this_Row.width
		//this_Row_height:=this_Row.height
		for _, device := range this_Row.devices {
			nThCropedFileName = strings.Replace(path.Base(mediaFile), path.Ext(mediaFile), "-"+strconv.Itoa(device.ID)+path.Ext(mediaFile), 1)
			//convert flag.png -crop 40x300+100+10  +repage  repage.gif
			cropMediaCommnd := exec.Command("convert", resizedFileName, "-verbose", "-crop", fmt.Sprintf("%vx%v+%v+%v", device.Width, device.Height, device.PrevWidth, this_Row.ordinate), "+repage", nThCropedFileName)
			cropMediaCommnd.Dir = "../Frontend/Media"
			fmt.Println(cropMediaCommnd.Args)
			output, err := cropMediaCommnd.CombinedOutput()
			if err != nil {
				log.Println("err", err)
			}
			fmt.Println(string(output))
		}
	}

}

func getTableData(tableOfDevice map[int]RowofDevice) ( int, int) {
	var width_table_Max int
	var height_table int
	for _, v := range tableOfDevice {
			if width_table_Max < v.width {
			width_table_Max = v.width
		}
		height_table += v.height
	}
	fmt.Printf("\n*****\nTable width_Max:%v,\t Height:%v\n*****\n", width_table_Max, height_table)
	return width_table_Max, height_table
}

// FindMatchingKnocks : sends matching knocks to chan confirmedKnockChanList and deletes it from recvdKnocks.Knocks
// TODO make sure to lock mutex recvdKnocks.MX.Lock() from where the function is called
func FindMatchingKnocks() {
	for key1, knock1 := range recvdKnocks.Knocks {
		for key2, knock2 := range recvdKnocks.Knocks {
			if knock1 != knock2 && knock1.Id != knock2.Id {
				t1 := knock1.TimeSend
				t2 := knock2.TimeSend
				delaybetweenKnocks := time.Duration(time.Second)
				if t1.Before(t2) {
					delaybetweenKnocks = t2.Sub(t1)
				} else {
					delaybetweenKnocks = t1.Sub(t2)
				}
				if delaybetweenKnocks < goodDurKnock {
					CKPCMux.Lock()
					ConfirmedKnockPairCount += 1

					knock1ChanG, found := confirmedKnockChanList.Load(knock1.Id)
					if !found {
						log.Println("knockedChan chan not found for ID", knock1.Id)
					}
					knock1Chan := knock1ChanG.(chan confirmedKnock)

					knock2ChanG, found := confirmedKnockChanList.Load(knock2.Id)
					if !found {
						log.Println("knockedChan chan not found for ID", knock2.Id)
					}
					knock2Chan := knock2ChanG.(chan confirmedKnock)

					knock1Chan <- confirmedKnock{knock1.Id, knock1.Loc, ConfirmedKnockPairCount}
					knock2Chan <- confirmedKnock{knock2.Id, knock2.Loc, ConfirmedKnockPairCount}
					//Adding neighbours to each other
					var thisKnock, NeighbourKnock knock = knock1, knock2

					this_deviceG, found := conctdDevices.Load(thisKnock.Id)
					if !found {
						log.Println("ERR device not found with ", thisKnock.Id)
					}
					this_device := this_deviceG.(*Device)

					Neigbour_deviceG, found := conctdDevices.Load(NeighbourKnock.Id)
					if !found {
						log.Println("ERR device not found with ", NeighbourKnock.Id)
					}
					Neigbour_device := Neigbour_deviceG.(*Device)

					this_device.AddNighbour(thisKnock.Loc, Neigbour_device)
					Neigbour_device.AddNighbour(NeighbourKnock.Loc, this_device)

					fmt.Println("After adding nighbours =================")
					printConctdDevices()

					CKPCMux.Unlock()

					//	Removing knock1 and knock2
					delete(recvdKnocks.Knocks, key1)
					delete(recvdKnocks.Knocks, key2)
				}
			}
		}

	}
}

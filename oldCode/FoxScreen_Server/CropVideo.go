//TODO export Crop()
package main

import (
	"fmt"
	"github.com/satori/go.uuid"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"errors"
	"io/ioutil"
)


const videoF = "test.mp4"
const OutputDir = "CropResult"

var wd string
var err error


func main() {
	Crop()
}
//TODO pass parameter to Crop()
func Crop() {
	//if _, err := os.Stat(OutputDir); os.IsNotExist(err) {
	//	os.MkdirAll(OutputDir, 0777) //ahhrgggg perm bits
	//}
	log.SetFlags(log.Lshortfile)
	fmt.Println("Starting to Crop :", videoF)
	wd, err = os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(wd)
	outFile, err := ffmpeg(videoF, 200, 200, 100, 100)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(outFile)
}

func ffmpeg(inVideo string, w int, h int, x int, y int) (string, error) {
	wc := strconv.Itoa(w)
	hc := strconv.Itoa(h)
	xc := strconv.Itoa(x)
	yc := strconv.Itoa(y)
	inVideoc := strings.Join([]string{wd, inVideo}, "/")
	outVideoc := strings.Join([]string{wd, OutputDir, uuid.NewV1().String() + ".mp4"}, "/")
	cmd := exec.Command("ffmpeg", "-loglevel", "fatal", "-i", inVideoc, "-filter:v", fmt.Sprintf("crop=%v:%v:%v:%v", wc, hc, xc, yc), outVideoc)
	fmt.Println("Running >>", cmd.Args, ", in Dir >>", cmd.Dir, ", in Path >>", cmd.Path)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Start()
	cmdOut, _ := ioutil.ReadAll(stdout)
	cmdErr, _ := ioutil.ReadAll(stderr)
	cmd.Wait()
	if len(cmdOut) > 0 {
		fmt.Println("STD out \n", string(cmdOut))
	}
	if len(cmdErr) > 0 {
		fmt.Println("err >>>", cmdErr)
		return "", errors.New(string(cmdErr))
	} else {
		return outVideoc, nil
	}
}

package main

import (
	"os/exec"
	"fmt"
	"log"
)

func main() {
	totalWidth:=1000
	totalHeight:=1000
	resizeMediaCommnd:=exec.Command("convert",
	resizeMediaCommnd.Dir="/home/usr1/projects/FoxScreen/Frontend/Media"
	fmt.Println(resizeMediaCommnd.Args)
	fmt.Println(resizeMediaCommnd.Dir)
	output,err:=resizeMediaCommnd.CombinedOutput()
	if err != nil {
		log.Println("err",err)
	}
	fmt.Println(string(output))

}

//
// teffa-cp 
//
// to receive files:
//     tcp

// to send files:
//     tcp <file> <ip/hostname>

// https://stackoverflow.com/questions/11268943/is-it-possible-to-capture-a-ctrlc-signal-and-run-a-cleanup-function-in-a-defe
package main

import (
	"fmt"
	"strconv"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const BUFFERSIZE = 1024
const MAXFILENAMESIZE = 512
const MAXFILESIZE = 15
const PROGRESSBARSIZE = 50

func byteToHuman(size int64) string {

	if size > 1024*1024*1024 {
		return(strconv.FormatFloat(float64(size)/(1024*1024*1024),'f',2,64) + "G")
	}else if size > 1024*1024 {
		return(strconv.FormatFloat(float64(size)/(1024*1024),'f',2,64) + "M")
        } else if size > 1024 {
                return(strconv.Itoa(int(size/1024)) + "K")
        }
	return strconv.Itoa(int(size))
}

func progressBar(received *int64, total *int64) {

	//fmt.Println("")
	for {
		fmt.Printf("\r")
		fmt.Printf("%s/%s [", byteToHuman(*received),byteToHuman(*total))
		for i := 0; i < int(float32(*received)/float32(*total) * PROGRESSBARSIZE); i++ {
			fmt.Printf("X")
		}
		for i := 0; i < int(PROGRESSBARSIZE - int(float32(*received)/float32(*total) * PROGRESSBARSIZE)); i++ {
			fmt.Printf(" ")
		}
		fmt.Printf("]    %d%%    ",int((float32(*received) / float32(*total))*100))
		time.Sleep(100 * time.Millisecond)
		if *received >= *total{
			break
		}
	}
}

func receiveFile(port string) {
	server, err := net.Listen("tcp", ":" + port)
	if err != nil {
		fmt.Println("Error listetning: ", err)
		os.Exit(1)
	}
	defer server.Close()
	fmt.Printf("Waiting for connections on port %s...\n",port)

	connection, err := server.Accept()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	// Buffer for name and size
	bufferFileName := make([]byte, MAXFILENAMESIZE)
	bufferFileSize := make([]byte, MAXFILESIZE)

	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")

	fileSizeH := byteToHuman(fileSize)

	fmt.Printf("Filename: %s\nSize: %s\n",fileName,fileSizeH)

	if _, err := os.Stat(fileName); err == nil {
		 fmt.Printf("*** file %s exists, it will be overwrited! ***\n",fileName)
	}

	var result string
	fmt.Printf("Accept the file (in the current dir) [y/N]: ")

	for {
		fmt.Scanf("%s",&result)
		if result == "y" || result == "yes" || result == "Y" || result == "YES" {
			// continue the code
			//TODO: Send a positive feedback
			connection.Write([]byte("Y"))
			break
		} else if result == "n" || result == "no" || result == "N" || result == "NO" {
			//TODO: Send a negative feedback
			connection.Write([]byte("N"))
			os.Exit(0)
		} else {
		        fmt.Printf("Accept the file (in the current dir) [y/N]: ")
		}
	}

	newFile, err := os.Create(fileName)

	if err != nil {
		panic(err)
	}
	defer newFile.Close()
	var receivedBytes int64

	fmt.Println("")
	go progressBar(&receivedBytes, &fileSize)

	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(newFile, connection, (fileSize - receivedBytes))
			connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
//			progressBar(receivedBytes,fileSize)
			break
		}
		io.CopyN(newFile, connection, BUFFERSIZE)
		receivedBytes += BUFFERSIZE
//		progressBar(receivedBytes, fileSize)
	}

	receivedBytes = fileSize
	progressBar(&receivedBytes, &fileSize)

	fmt.Println("\nReceived file completely!")

}

func sendFile(filename string, ip string) {

	if len(filename) > MAXFILENAMESIZE {
		fmt.Printf("Error: Filename size > %d!\n",MAXFILENAMESIZE)
		os.Exit(1)
	}
	fmt.Printf("Sending the file: %s to %s, Waiting for accept...\n",filename, ip)
	connection, err := net.Dial("tcp", ip +":2000")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer connection.Close()
//	fmt.Println("Connected to server, start sending the file name and file size")
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(strconv.FormatInt(fileInfo.Size(), 10)) > MAXFILESIZE {
	  fmt.Printf("Error: Filesize size > %d!\n",MAXFILESIZE)
                os.Exit(1)
        }

	fileSizeInt := fileInfo.Size()
	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), MAXFILESIZE)
	fileName := fillString(fileInfo.Name(), MAXFILENAMESIZE)
//	fmt.Println("Sending filename and filesize!")
	connection.Write([]byte(fileSize))
	connection.Write([]byte(fileName))

	//TODO: listen for a feedback, before sending the file
	fileAccepted := make([]byte, 1)
	connection.Read(fileAccepted)

	if string(fileAccepted) == "N" {
		fmt.Println("File not accepted!")
		os.Exit(0)
	}

	sendBuffer := make([]byte, BUFFERSIZE)
	fmt.Printf("Start sending file!\n\n")
	var sentBytes int64
	go progressBar(&sentBytes, &fileSizeInt)
	for {
		num, err := file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
		sentBytes+=int64(num)
	}

	progressBar(&sentBytes, &fileSizeInt)

	fmt.Println("File has been sent, closing connection!")
	return
}

func fillString(returnString string, toLength int) string {
	for {
		lengtString := len(returnString)
		if lengtString < toLength {
			returnString = returnString + ":"
			continue
		}
		break
	}
	return returnString
}

func main() {

	if len(os.Args) < 2 {
		// no args (receive mode)
		receiveFile("2000")
	} else if len(os.Args) == 3 {
		// 2 args (send mode)
		// <file> <ip>
		sendFile(os.Args[1], os.Args[2])
	} else {
		fmt.Printf("Wrong number of args (%d), expected 0 (to receive files) or 2 (to send files)\n",len(os.Args)-1)
		fmt.Print("Usage:\n")
		fmt.Print("    tcp\t\t\t\t: to receive files\n")
		fmt.Print("    tcp <file> <ip/hostname>\t: to send files\n")
		os.Exit(1)
	}

}

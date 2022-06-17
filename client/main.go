package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	helloworldPb "github.com/mwqnice/proto/helloworld"
	"google.golang.org/grpc"
	"io"
	"log"
	"strconv"
	"sync"
	"time"
)

func main()  {

	r := gin.Default()
	r.GET("/get_data", GetData)
	r.GET("/get_stream", GetStream)
	r.GET("/set_stream", SetStream)
	r.GET("/all_stream", AllStream)
	r.Run(":8090")
}

func GetGrpcCon() (*grpc.ClientConn, error){
	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return conn, err
}

//模拟一元RPC
func GetData(c *gin.Context)  {
	//创建微服务客户端
	con, err := GetGrpcCon()
	if err != nil {
		fmt.Println("get-grpc-con err:", err)
		return
	}
	client := helloworldPb.NewHelloworldClient(con)
	rep, err :=client.GetData(context.Background(), &helloworldPb.ReqData{Name: "test"})
	if err != nil {
		fmt.Println("client get-data err:", err)
		return
	}
	c.JSON(200, rep.Reply)
}

//模拟服务端流式 RPC
func GetStream(c *gin.Context)  {
	//创建微服务客户端
	con, err := GetGrpcCon()
	if err != nil {
		fmt.Println("get-grpc-con err:", err)
		return
	}
	client := helloworldPb.NewHelloworldClient(con)
	rep, err := client.GetStream(context.Background(), &helloworldPb.ReqData{Name: "test"})
	if err != nil {
		fmt.Println("client get-stream err:", err)
		return
	}
	for  {
		repData, err := rep.Recv()
		if err != nil {
			if err != io.EOF {
				fmt.Println("服务端发送完毕！")
				break
			}
			fmt.Println("服务端发送错误：", err)
			break
		}
		fmt.Printf("接收服务端数据：%v",repData)
	}
}

//模拟客户端流式 RPC
func SetStream(c *gin.Context)  {
	//创建微服务客户端
	con, err := GetGrpcCon()
	if err != nil {
		fmt.Println("get-grpc-con err:", err)
		return
	}
	client := helloworldPb.NewHelloworldClient(con)
	rstream, err := client.SetStream(context.Background())
	if err != nil {
		fmt.Println("client set-stream err:", err)
		return
	}
	if rstream == nil {
		fmt.Println("stream is nil")
		return
	}
	//发送数据
	var sendData = []string{"张三", "李四", "王五"}
	for _, v := range sendData {
		err := rstream.Send(&helloworldPb.ReqData{Name: v})
		if err != nil {
			fmt.Println("客户端发送错误：", err)
			break
		}
		time.Sleep(time.Second)
	}
	//接收数据
	repData, err := rstream.CloseAndRecv()
	if err != nil {
		fmt.Println("接收数据错误：", err)
		return
	}
	c.JSON(200, repData.Reply)
}

//模拟双向流式 RPC
func AllStream(c *gin.Context)  {
	//创建微服务客户端
	con, err := GetGrpcCon()
	if err != nil {
		fmt.Println("get-grpc-con err:", err)
		return
	}
	client := helloworldPb.NewHelloworldClient(con)
	stream, err := client.AllStream(context.Background())
	if err != nil {
		fmt.Println("all-stream err:", err)
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		i := 0
		for {
			if err = stream.Send(&helloworldPb.ReqData{Name: "client" + strconv.Itoa(i)}); err != nil {
				if err == io.EOF {
					log.Println("")
				} else {
					log.Fatalf("客户端发送失败, err:%#v", err.Error())
				}
				return
			}

			if i >10 {
				if err = stream.CloseSend(); err != nil {
					log.Fatalf("客户端关闭发送流失败, err:%#v", err.Error())
				} else {
					log.Println("客户端发送完成")
				}
				return
			}
			i++
		}
	}()

	go func() {
		defer wg.Done()
		var repData *helloworldPb.RepData
		for  {
			if repData, err = stream.Recv(); err != nil {
				if err == io.EOF {
					log.Fatalf("客户端接收失败, err:%#v", err.Error())
				} else {
					log.Fatalf("客户端接收失败, err:%#v", err.Error())
				}
			}
			fmt.Printf("客户端接收数据：%#v\n", repData.Reply)
		}
	}()

	wg.Wait()
}
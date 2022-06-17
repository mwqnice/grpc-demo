package main

import (
	"context"
	helloworldPb "github.com/mwqnice/proto/helloworld"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

type Helloworld struct {}

func (h *Helloworld) GetData(ctx context.Context, req *helloworldPb.ReqData) (*helloworldPb.RepData, error) {
	return &helloworldPb.RepData{Reply:"你好：" + req.Name, }, nil
}
func (h *Helloworld) GetStream(req *helloworldPb.ReqData, stream helloworldPb.Helloworld_GetStreamServer) error {
	i := 1
	for  {
		if i > 10 {
			break
		}
		err := stream.Send(&helloworldPb.RepData{
			Reply: "你好："+req.Name+ strconv.Itoa(i),
		})
		if err != nil {
			log.Fatal("send to client error:", err.Error())
		}
		time.Sleep(time.Second)
		i++
	}

	return nil
}
func (h *Helloworld) SetStream(req helloworldPb.Helloworld_SetStreamServer) error {
	for  {
		reqData, err := req.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("服务端处理完毕" )
				break
			} else {
				log.Fatalf("服务端接收错误,err:%#v", err.Error())
			}
		} else {
			log.Printf("服务端接收的数据是：%#v\n", reqData)
		}
	}
	if err := req.SendMsg(&helloworldPb.RepData{
		Reply: "GoodBye",
	}); err != nil {
		log.Fatalf("服务端发送数据错误,err:%#v", err.Error())
	}
	return nil
}
func (h *Helloworld) AllStream(req helloworldPb.Helloworld_AllStreamServer) error {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			reqData, err := req.Recv()
			if err != nil {
				if err == io.EOF {
					log.Println("服务端处理完毕" )
					break
				} else {
					log.Fatalf("服务端接收错误,err:%#v", err.Error())
				}
			} else {
				log.Printf("服务端接收的数据是：%#v\n", reqData.Name)
			}
		}
	}()

	go func() {
		defer wg.Done()
		i := 0
		for {
			if err := req.Send(&helloworldPb.RepData{Reply: "服务端发送数据:" + strconv.Itoa(i)}); err != nil {
				log.Fatalf("服务端发送数据失败, err:%#v", err.Error())
			}

			if i > 5 {
				break
			}

			i++
			time.Sleep(time.Second)
		}
	}()

	wg.Wait()

	return nil
}

func main()  {
	server := grpc.NewServer()
	//注册
	helloworldPb.RegisterHelloworldServer(server, new(Helloworld))

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err.Error())
	}
	server.Serve(lis)
}
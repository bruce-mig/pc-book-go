package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gitlab.com/brucemig/pcbook/client"
	"gitlab.com/brucemig/pcbook/pb"
	"gitlab.com/brucemig/pcbook/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	refreshDuration = 30 * time.Second
)

func main() {
	viper.SetConfigName(".env") // name of the config file (without extension)
	viper.AddConfigPath(".")    // path to look for the config file in
	viper.SetConfigType("env")  // REQUIRED if the config file does not have the extension in the name
	viper.AutomaticEnv()        // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Fatalf("Error while reading config file %s", err)
	}

	// for _, key := range viper.AllKeys() {
	// 	fmt.Printf("Key: %s, Value: %s\n", key, viper.GetString(key))
	// }

	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddress)

	cc1, err := grpc.Dial(*serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("cannot dial server:", err)
	}

	authClient := client.NewAuthClient(cc1, viper.GetString("USERNAME1"), viper.GetString("PASSWORD"))
	interceptor, err := client.NewAuthInterceptor(authClient, authMethods(), viper.GetDuration("REFRESH_DURATION")*time.Second)
	if err != nil {
		log.Fatal("cannot create new interceptor: ", err)
	}

	cc2, err := grpc.Dial(*serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Fatal("cannot dial server:", err)
	}

	laptopClient := client.NewLaptopClient(cc2)
	testRateLaptop(laptopClient)

}

func authMethods() map[string]bool {
	const laptopServicePath = "/brucemig.pcbook.LaptopService/"
	return map[string]bool{
		laptopServicePath + "CreateLaptop": true,
		laptopServicePath + "UploadImage":  true,
		laptopServicePath + "RateLaptop":   true,
	}
}

func testCreateLaptop(laptopClient *client.LaptopClient) {
	laptopClient.CreateLaptop(sample.NewLaptop())
}

func testSearchLaptop(laptopClient *client.LaptopClient) {
	for i := 0; i < 10; i++ {
		laptopClient.CreateLaptop(sample.NewLaptop())
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	laptopClient.SearchLaptop(filter)
}

func testUploadImage(laptopClient *client.LaptopClient) {
	laptop := sample.NewLaptop()
	laptopClient.CreateLaptop(laptop)
	laptopClient.UploadImage(laptop.GetId(), "tmp/laptop.jpeg")
}

func testRateLaptop(laptopClient *client.LaptopClient) {
	n := 3
	laptopIDs := make([]string, n)

	for i := 0; i < n; i++ {
		laptop := sample.NewLaptop()
		laptopIDs[i] = laptop.GetId()
		laptopClient.CreateLaptop(laptop)
	}

	scores := make([]float64, n)
	for {
		fmt.Printf("rate laptop (y/n)?")
		var answer string
		fmt.Scan(&answer)

		if strings.ToLower(answer) != "y" {
			break
		}

		for i := 0; i < n; i++ {
			scores[i] = sample.RandomLaptopScore()
		}

		err := laptopClient.RateLaptop(laptopIDs, scores)
		if err != nil {
			log.Fatal(err)
		}

	}
}

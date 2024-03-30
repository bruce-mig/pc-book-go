package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/spf13/viper"
	"gitlab.com/brucemig/pcbook/pb"
	"gitlab.com/brucemig/pcbook/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

func seedUsers(userStore service.UserStore) error {
	if err := createUser(
		userStore,
		viper.GetString("USERNAME1"),
		viper.GetString("PASSWORD"),
		viper.GetString("ROLE1"),
	); err != nil {
		return err
	}

	return createUser(
		userStore,
		viper.GetString("USERNAME2"),
		viper.GetString("PASSWORD"),
		viper.GetString("ROLE2"),
	)
}

func createUser(userStore service.UserStore, username, password, role string) error {

	user, err := service.NewUser(username, password, role)
	if err != nil {
		return fmt.Errorf("error creating user %s: %v", username, err)
	}

	err = userStore.Save(user)
	if err != nil {
		return fmt.Errorf("error saving user %s: %v", username, err)
	}

	return nil
}

func accessibleRoles() map[string][]string {
	const laptopServicePath = "/brucemig.pcbook.LaptopService/"
	return map[string][]string{
		laptopServicePath + "CreateLaptop": {viper.GetString("ROLE1")},
		laptopServicePath + "UploadImage":  {viper.GetString("ROLE1")},
		laptopServicePath + "RateLaptop":   {viper.GetString("ROLE1"), viper.GetString("ROLE2")},
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed client's certificate
	pemClientCA, err := os.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair("cert/server-cert.pem", "cert/server-key.pem")
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}

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

	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("server listening on port %d", *port)

	userStore := service.NewInMemoryUserStore()
	if err := seedUsers(userStore); err != nil {
		log.Fatal("cannot seed users:", err)
	}

	jwtManager := service.NewJWTManager(viper.GetString("SECRET_KEY"), viper.GetDuration("TOKEN_DURATION")*time.Minute)
	authServer := service.NewAuthServer(userStore, jwtManager)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")
	ratingStore := service.NewInMemoryRatingStore()

	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal("cannot load TLS credentials: ", err)
	}

	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
	grpcServer := grpc.NewServer(
		grpc.Creds(tlsCredentials),
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	)

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	reflection.Register(grpcServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

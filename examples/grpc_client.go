package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "github.com/imnitish-dev/ip2location/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Command line flags
	ip := flag.String("ip", "", "IP address to lookup")
	server := flag.String("server", "localhost:50051", "gRPC server address")
	timeout := flag.Duration("timeout", 5*time.Second, "Timeout for request")
	flag.Parse()

	// Check if IP is provided
	if *ip == "" {
		log.Fatal("Please provide an IP address using -ip flag")
	}

	// Connect to gRPC server
	conn, err := grpc.Dial(*server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := pb.NewIP2LocationServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Make the request
	resp, err := client.LookupIP(ctx, &pb.LookupRequest{Ip: *ip})
	if err != nil {
		log.Fatalf("Could not lookup IP: %v", err)
	}

	// Print response in a formatted way
	fmt.Println("\nIP Lookup Results:")
	fmt.Println("==================")
	
	if resp.Message != "" {
		fmt.Printf("Message: %s\n", resp.Message)
	}

	if resp.Maxmind != nil {
		fmt.Println("\nMaxMind Data:")
		fmt.Printf("  Country:      %s\n", resp.Maxmind.Country)
		fmt.Printf("  City:         %s\n", resp.Maxmind.City)
		fmt.Printf("  Region:       %s\n", resp.Maxmind.Region)
		fmt.Printf("  Country Code: %s\n", resp.Maxmind.CountryCode)
		fmt.Printf("  Coordinates:  %.6f, %.6f\n", resp.Maxmind.Latitude, resp.Maxmind.Longitude)
	}

	if resp.Ip2Location != nil {
		fmt.Println("\nIP2Location Data:")
		fmt.Printf("  Country:      %s\n", resp.Ip2Location.Country)
		fmt.Printf("  City:         %s\n", resp.Ip2Location.City)
		fmt.Printf("  Region:       %s\n", resp.Ip2Location.Region)
		fmt.Printf("  Country Code: %s\n", resp.Ip2Location.CountryCode)
		fmt.Printf("  Coordinates:  %.6f, %.6f\n", resp.Ip2Location.Latitude, resp.Ip2Location.Longitude)
	}
} 
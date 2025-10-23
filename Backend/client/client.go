package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pb "backend/hazard"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewHazardDetectionClient(conn)

	// Read image file
	imgPath := "../images/img.jpeg"
	imgData, err := os.ReadFile(imgPath)
	if err != nil {
		log.Fatalf("could not read image file: %v", err)
	}

	// Create request
	req := &pb.ImageRequest{
		ImageData: imgData,
		Latitude:  34.0522,
		Longitude: -118.2437,
	}

	// Call RPC
	res, err := c.DetectHazard(context.Background(), req)
	if err != nil {
		log.Fatalf("could not detect hazard: %v", err)
	}

	fmt.Printf("Response: %v\n", res)
}

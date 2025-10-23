package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net"
	"os"

	pb "backend/hazard"

	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/ollama/ollama/api"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedHazardDetectionServer
}

// DetectHazard receives image bytes and returns a dummy detection
func (s *server) DetectHazard(ctx context.Context, req *pb.ImageRequest) (*pb.DetectionResponse, error) {
	fmt.Printf("Received image of size: %d bytes\n", len(req.GetImageData()))
	fmt.Printf("Location: %f, %f\n", req.GetLatitude(), req.GetLongitude())
	resp, err := detectHazardWithOllama(ctx, req.GetImageData())
	if err != nil {
		return nil, err
	}
	fmt.Println("Response : ", resp)
	return resp, nil
}
func serveFrames(imgByte []byte) {

	img, _, err := image.Decode(bytes.NewReader(imgByte))
	if err != nil {
		log.Fatalln(err)
	}

	out, _ := os.Create("./images/img.jpeg")
	defer out.Close()

	var opts jpeg.Options
	opts.Quality = 100

	err = jpeg.Encode(out, img, &opts)
	//jpeg.Encode(out, img, nil)
	if err != nil {
		log.Println(err)
	}

}
func detectHazardWithOllama(ctx context.Context, imgData []byte) (*pb.DetectionResponse, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	format := json.RawMessage(`"json"`)
	req := &api.GenerateRequest{
		Model:  "llava",
		Prompt: "Analyse the image and identify the hazard, confidence score, and priority level. Return the response as a JSON object with the fields: hazard_type, confidence, and priority. hazard_type is in string , confidence is in float upto 0 to 100, priority is in int upto 1 to 3 (1 == low, 2 == medium,3==high.If the image is not about road hazard then hazard_type should be empty or '' and confidence 100 ,priority 0",
		Images: []api.ImageData{imgData},
		Format: format,
	}
	var fullResponse string
	resp := &pb.DetectionResponse{}
	fmt.Println("Ollama is running !!")
	var ollamaResp struct {
		HazardType string  `json:"hazard_type"`
		Confidence float32 `json:"confidence"`
		Priority   int32   `json:"priority"`
	}
	err = client.Generate(ctx, req, func(res api.GenerateResponse) error {
		fullResponse += res.Response
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(fullResponse), &ollamaResp); err != nil {
		return nil, err
	}
	resp = &pb.DetectionResponse{
		HazardType: ollamaResp.HazardType,
		Confidence: ollamaResp.Confidence,
		Priority:   ollamaResp.Priority,
	}

	return resp, nil
}

func main() {
	port := ":8080"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterHazardDetectionServer(grpcServer, &server{})

	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrappedGrpc.ServeHTTP(w, r)
	})

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "X-User-Agent", "X-Grpc-Web"},
	}).Handler(handler)

	httpServer := &http.Server{
		Addr:    port,
		Handler: corsHandler,
	}

	fmt.Println("Go gRPC server running on ", port)
	if err := httpServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

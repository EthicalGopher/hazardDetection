package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	pb "backend/hazard"

	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/joho/godotenv"
	"github.com/ollama/ollama/api"
	"github.com/rs/cors"
	"google.golang.org/genai"
	"google.golang.org/grpc"
)

var prompt = `
Analyze the given image and detect any visible road hazards such as potholes, speed breakers, animal crossings, fallen trees, debris, waterlogging, or damaged road sections.
Return the result in JSON format with the following fields:

hazard_type (string) — the detected type of road hazard (e.g., "pothole", "speed breaker", "animal crossing").

confidence (float, 0–100) — model’s confidence level in the detection.

priority (int, 1–3) — urgency level of the hazard, where

1 = low,

2 = medium,

3 = high.

Example Output:

{
"hazard_type": "pothole",
"confidence": 92.7,
"priority": 3
}
`

type server struct {
	pb.UnimplementedHazardDetectionServer
	ollamaClient *api.Client
}

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}
}

// DetectHazard receives image bytes and returns a dummy detection
func (s *server) DetectHazard(ctx context.Context, req *pb.ImageRequest) (*pb.DetectionResponse, error) {
	fmt.Printf("Received image of size: %d bytes\n", len(req.GetImageData()))
	fmt.Printf("Location: %f, %f\n", req.GetLatitude(), req.GetLongitude())
	//serveFrames(req.GetImageData())
	//resp := hazardDetectionWithGemini()
	resp, err := s.detectHazardWithOllama(ctx, req.GetImageData())
	if err != nil {
		log.Println(err)
		return nil, err
	}
	fmt.Println("Response : ", resp)
	return resp, nil
}
func (s *server) detectHazardWithOllama(ctx context.Context, imgData []byte) (*pb.DetectionResponse, error) {
	format := json.RawMessage(`"json"`)
	req := &api.GenerateRequest{
		Model:  "llava",
		Prompt: prompt,
		Format: format,
		Images: []api.ImageData{imgData},
	}
	var fullResponse string
	resp := &pb.DetectionResponse{}
	fmt.Println("Ollama is running !!")
	var ollamaResp struct {
		HazardType string  `json:"hazard_type"`
		Confidence float32 `json:"confidence"`
		Priority   int32   `json:"priority"`
	}
	err := s.ollamaClient.Generate(ctx, req, func(res api.GenerateResponse) error {
		fullResponse += res.Response
		return nil
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("Raw JSON from Ollama:", fullResponse)
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

func hazardDetectionWithGemini() *pb.DetectionResponse {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"hazard_type": {Type: genai.TypeString},
				"confidence":  {Type: genai.TypeNumber},
				"priority":    {Type: genai.TypeInteger},
			},
			Required: []string{"hazard_type", "confidence", "priority"},
		},
	}
	bytes, _ := os.ReadFile("./images/img.jpeg")

	parts := []*genai.Part{
		genai.NewPartFromBytes(bytes, "image/jpeg"),
		genai.NewPartFromText(prompt),
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		contents,
		config,
	)
	if err != nil {
		log.Fatal(err)
	}

	var geminiResp struct {
		HazardType string  `json:"hazard_type"`
		Confidence float32 `json:"confidence"`
		Priority   int32   `json:"priority"`
	}
	jsonString := result.Candidates[0].Content.Parts[0].Text
	err = json.Unmarshal([]byte(jsonString), &geminiResp)
	if err != nil {
		fmt.Println(err)
	}
	return &pb.DetectionResponse{
		HazardType: geminiResp.HazardType,
		Confidence: geminiResp.Confidence,
		Priority:   geminiResp.Priority,
	}
}

func main() {
	ollamaClient, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal("failed to create ollama client: %v", err)
	}
	port := ":8080"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterHazardDetectionServer(grpcServer, &server{ollamaClient: ollamaClient})

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

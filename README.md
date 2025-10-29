# Hazard Detection System

This project is a full-stack application designed to detect hazards in a video stream and alert the user. It uses a Go backend with gRPC for video processing and a React frontend for displaying the video and hazard alerts.

## Features

*   Real-time hazard detection in video streams.
*   Frontend built with React and Vite for a fast and modern user experience.
*   Backend powered by Go for efficient video processing.
*   gRPC for communication between the frontend and backend.
*   Select between internal and external camera feeds.

## Technologies Used

**Frontend:**

*   React
*   Vite
*   Tailwind CSS
*   gRPC-web
*   Material-UI

**Backend:**

*   Go
*   gRPC
*   Ollama
*   Google GenAI

## Project Structure

```
hazardDetection/
├── Backend/
│   ├── main.go         # Main application entry point
│   ├── hazard/         # gRPC generated files
│   ├── client/         # gRPC client
│   └── ...
└── Frontend/
    ├── src/
    │   ├── App.jsx       # Main React component
    │   ├── components/   # React components
    │   └── proto/        # gRPC generated files
    └── ...
```

## Getting Started

### Prerequisites

*   Go (version 1.20+)
*   Node.js (version 18+)
*   Protocol Buffers Compiler (`protoc`)
*   `protoc-gen-ts` plugin for TypeScript generation

### Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/your-username/hazard-detection.git
    cd hazard-detection
    ```

2.  **Backend Setup:**

    ```bash
    cd Backend
    # Create a .env file and add your Google API Key
    echo "GOOGLE_API_KEY=your_api_key" > .env
    go mod tidy
    # Generate gRPC code
    protoc --go_out=. --go-grpc_out=. hazard/hazard.proto
    ```

3.  **Frontend Setup:**

    ```bash
    cd Frontend
    npm install
    ```

### Running the Application

1.  **Start the Backend Server:**

    ```bash
    cd Backend
    go run main.go
    ```

2.  **Start the Frontend Development Server:**

    ```bash
    cd Frontend
    npm run dev
    ```

## Usage

1.  Open your browser and navigate to the address provided by Vite (usually `http://localhost:5173`).
2.  Select either the internal or external camera feed.
3.  The video stream will be displayed, and any detected hazards will be highlighted.

## API Reference

The application uses a gRPC API for communication between the frontend and backend. The API is defined in the `hazard.proto` file.

*   **`HazardDetection` Service:**
    *   `DetectHazards`: A streaming RPC that sends video frames from the client to the server and receives hazard detection results.


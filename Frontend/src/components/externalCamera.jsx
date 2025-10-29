import React, { useState, useEffect, useRef, useCallback } from "react";
import { HazardDetectionClient} from "../proto/hazardPackage/hazard.client"
import { GrpcWebFetchTransport } from "@protobuf-ts/grpcweb-transport";
import { ImageRequest } from "../proto/hazardPackage/hazard"
import { MapPinned,RotateCcw  } from 'lucide-react';
import Alert from '@mui/material/Alert';


// --- Helper Components ---

const Icon = ({ path, className = "h-8 w-8" }) => (
  <svg xmlns="http://www.w3.org/2000/svg" className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={path} />
  </svg>
);

const TooltipButton = ({ tooltip, children, ...props }) => (
  <div className="has-tooltip">
    <span className="tooltip rounded bg-gray-800 p-2 text-xs mt-10">{tooltip}</span>
    <button {...props}>{children}</button>
  </div>
);

// --- Main Application Component ---

const warningSound = new Audio('/warning.mp3');
const errorSound = new Audio('/error.mp3');

export default function ExternalCamera({ipWeb}) {
  // --- State and Refs ---
  const imgRef = useRef(null);
  const canvasRef = useRef(null);
  const [hazardPriority,setHazardPriority] = useState({
    "hazard_type":"",
    "priority":0,
    "confidence":100.00,
  });
  const [isMenuOpen, setMenuOpen] = useState(true);
  const [isRecording, setIsRecording] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [stopRecordingOnResponse, setStopRecordingOnResponse] = useState(false);


  const [status, setStatus] = useState("Idle");
  const transport = new GrpcWebFetchTransport({baseUrl:"http://localhost:8080"})
  const client = new HazardDetectionClient(transport)
  
  const captureAndSendFrame = useCallback(async () => {
    if (!imgRef.current || !canvasRef.current) {
      setStatus("Image not ready");
      return;
    }

    const img = imgRef.current;
    const canvas = canvasRef.current;
    const context = canvas.getContext("2d");

    try {
      canvas.width = img.width;
      canvas.height = img.height;
      context.drawImage(img, 0, 0, canvas.width, canvas.height);

      const frame = canvas.toDataURL("image/jpeg", 0.6).split(',')[1]; // 60% quality
      console.log("frame data length:", frame.length);
      
      const gpsLat = 26.15; // Placeholder
      const gpsLong = 91.77; // Placeholder

      try {
        setIsSending(true);
        setStatus("Sending frame...");
        const binaryFrame = new Uint8Array(atob(frame).split('').map(char => char.charCodeAt(0)))
        const request = ImageRequest.create({
          imageData : binaryFrame,
          latitude:gpsLat,
          longitude:gpsLong
        })
        const {response} = await client.detectHazard(request)
        setHazardPriority({
          "hazard_type":response.hazardType,
          "priority":response.priority,
          "confidence":response.confidence
        })
        console.log(response)
        setStatus("Frame sent successfully!");
        if (stopRecordingOnResponse) {
          setIsRecording(false);
          setStopRecordingOnResponse(false);
        }
      } catch (err) {
        console.error("Error sending frame:", err);
        setStatus(`Error: ${err.message}`);
      } finally {
        setIsSending(false);
      }
    } catch (e) {
      setStatus("Error capturing frame from image.");
      console.error(e);
    }
  }, [stopRecordingOnResponse]);

  useEffect(() => {
    let isCancelled = false;

    const sendFrameLoop = async () => {
      if (isRecording && !isCancelled) {
        await captureAndSendFrame();
        sendFrameLoop();
      }
    };

    sendFrameLoop();

    return () => {
      isCancelled = true;
    };
  }, [isRecording, captureAndSendFrame]);

  useEffect(() => {
    const playSound = async (sound) => {
      try {
        await sound.play();
      } catch (err) {
        console.error("Error playing sound:", err);
      }
    };

    if (hazardPriority.priority === 2) {
      playSound(warningSound);
    } else if (hazardPriority.priority === 3) {
      playSound(errorSound);
    }
  }, [hazardPriority]);




  // --- Render ---

  return (
    <>
    <div className="relative h-screen w-screen bg-black text-white font-sans">
      <div className="absolute top-0 flex justify-center items-center w-screen p-11">
    <div className="w-1/2">
        {hazardPriority.priority ==2 && (
      <Alert severity="warning">Hazard detected</Alert>
    )}
            {hazardPriority.priority == 3 && (
            <Alert severity="error">Critical Hazard detected {hazardPriority.hazard_type}</Alert>
            )}
        </div>
      </div>
      <img ref={imgRef} src = {ipWeb + "/video"} className="h-full w-full object-cover" crossOrigin="anonymous" />
      <canvas ref={canvasRef} className="hidden" />

      {/* Overlays */}
      <div className="absolute inset-0 flex flex-col justify-between p-4">
        {/* Top Section: Menu and Recording Indicator */}
        <div className="flex justify-between items-start">
          <div>
            <TooltipButton tooltip="Toggle Menu" onClick={() => setMenuOpen(!isMenuOpen)} className="text-white bg-black bg-opacity-30 backdrop-blur-md rounded-lg p-2">
              <Icon path={isMenuOpen ? "M6 18L18 6M6 6l12 12" : "M4 6h16M4 12h16m-7 6h7"} />
            </TooltipButton>
            <div
              className={`bg-black bg-opacity-30 backdrop-blur-md rounded-lg transition-all duration-300 ease-in-out mt-2 ${isMenuOpen ? "p-4" : "p-0 h-0 w-0 overflow-hidden"}`}>
              <h2 className={`text-lg font-bold transition-opacity ${isMenuOpen ? "opacity-100" : "opacity-0"}`}>Controls</h2>
              <div className={`flex justify-center items-center space-x-4 mt-4 transition-opacity ${isMenuOpen ? "opacity-100" : "opacity-0"}`}>
                <TooltipButton tooltip={isRecording ? "Stop Sending" : "Start Sending"} onClick={() => { if (isRecording) { setStopRecordingOnResponse(true); } else { setIsRecording(true); } }} className={isRecording ? "text-red-500" : "text-green-400"} disabled={isSending}>
                  <Icon path={isRecording ? "M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" : "M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"} />
                </TooltipButton>
                <TooltipButton tooltip="Not Implemented" disabled className="text-gray-500 cursor-not-allowed">
                  <RotateCcw/>
                </TooltipButton>
                <TooltipButton tooltip="Not Implemented" disabled className="text-gray-500 cursor-not-allowed">
                  <Icon path="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </TooltipButton>
                <TooltipButton tooltip="Not Implemented" disabled className="text-gray-500 cursor-not-allowed">
                  <MapPinned/>
                </TooltipButton>
              </div>
            </div>
          </div>

          {isRecording && (
            <div className="flex items-center space-x-2 bg-black bg-opacity-50 p-2 rounded-lg">
              <div className="h-3 w-3 rounded-full bg-red-500 animate-pulse"></div>
              <p className="text-sm font-bold">REC</p>
            </div>
          )}
        </div>

        {/* Bottom Section: Status Display */}
        <div className="bg-black bg-opacity-50 p-2 rounded-lg self-start">
          <p className="text-sm">Status: {status}</p>
        </div>
      </div>
    </div>
    </>
  );
}

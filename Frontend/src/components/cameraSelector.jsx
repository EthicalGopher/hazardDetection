
import React, { useState, useEffect } from 'react';
import ExternalCamera from './externalCamera';
import InternalCamera from './internalCamera';

export default function CameraSelector() {
  const [ipWeb, setIpWeb] = useState('');
  const [useInternal, setUseInternal] = useState(false);
  const [showInput, setShowInput] = useState(true);

  useEffect(() => {
    const storedIp = localStorage.getItem('ipWeb');
    if (storedIp) {
      setIpWeb(storedIp);
      setShowInput(false);
    }
  }, []);

  const handleIpSubmit = (e) => {
    e.preventDefault();
    localStorage.setItem('ipWeb', ipWeb);
    setShowInput(false);
  };

  const handleUseInternal = () => {
    setUseInternal(true);
    setShowInput(false);
  };

  const handleChangeIp = () => {
    setShowInput(true);
  };

  if (showInput) {
    return (
      <div className="h-screen w-screen bg-gray-800 flex flex-col justify-center items-center text-white">
        <h1 className="text-3xl mb-4">Select Camera</h1>
        <form onSubmit={handleIpSubmit} className="flex flex-col items-center">
          <input
            type="text"
            value={ipWeb}
            onChange={(e) => setIpWeb(e.target.value)}
            placeholder="Enter External Camera IP"
            className="p-2 rounded bg-gray-700 border border-gray-600 mb-4 w-64"
          />
          <button type="submit" className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded mb-2">
            Use External Camera
          </button>
        </form>
        <button onClick={handleUseInternal} className="bg-gray-500 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded">
          Use Internal Camera
        </button>
      </div>
    );
  }

  if (useInternal) {
    return <InternalCamera />;
  }

  return (
    <div>
      <ExternalCamera ipWeb={ipWeb} />
      <button onClick={handleChangeIp} className="absolute bottom-4 right-4 bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
        Change IP
      </button>
    </div>
  );
}

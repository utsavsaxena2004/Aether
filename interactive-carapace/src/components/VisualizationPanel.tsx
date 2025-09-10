'use client';

import React, { useState, useRef } from 'react';
import { BarChart3, Activity, Upload, Database, Square, PieChart, LineChart, BarChart, AlertCircle, CheckCircle } from 'lucide-react';
import { ChartRenderer } from './ChartRenderer';
import { useActiveChart, useChartHistory, useUploadedData, useIsSelectionMode, useSelectedArea, useChatStore } from '@/store/chatStore';
import { useWebSocket } from '@/hooks/useWebSocket';

export function VisualizationPanel() {
  const [dragOver, setDragOver] = useState(false);
  const [uploadStatus, setUploadStatus] = useState<{status: 'idle' | 'uploading' | 'success' | 'error', message: string}>({status: 'idle', message: ''});
  const activeChart = useActiveChart();
  const chartHistory = useChartHistory();
  const uploadedData = useUploadedData();
  const isSelectionMode = useIsSelectionMode();
  const selectedArea = useSelectedArea();
  const { sendVisualQuery } = useWebSocket();
  const [visualQuery, setVisualQuery] = useState('');
  const [showQueryInput, setShowQueryInput] = useState(false);
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    
    const files = Array.from(e.dataTransfer.files);
    if (files.length > 0) {
      handleFileUpload(files[0]);
    }
  };

  const handleFileUpload = async (file: File) => {
    // Validate file type
    if (!file.name.endsWith('.csv')) {
      setUploadStatus({status: 'error', message: 'Please upload a CSV file'});
      return;
    }

    // Validate file size (max 50MB)
    if (file.size > 50 * 1024 * 1024) {
      setUploadStatus({status: 'error', message: 'File size exceeds 50MB limit'});
      return;
    }

    setUploadStatus({status: 'uploading', message: 'Uploading file...'});

    try {
      // Create FormData and send to backend
      const formData = new FormData();
      formData.append('file', file);
      
      const response = await fetch('/upload', {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        throw new Error(`Upload failed with status ${response.status}`);
      }

      const data = await response.json();
      
      // Set uploaded data in store
      const { setUploadedData } = useChatStore.getState();
      setUploadedData({
        fileName: file.name,
        fileSize: file.size,
        fileType: file.type,
        rowCount: data.rowCount || 0,
        columnCount: data.columnCount || 0,
        columns: data.columns || [],
        sampleData: data.sampleData || [],
      });
      
      setUploadStatus({status: 'success', message: 'File uploaded successfully!'});
      
      // Clear status after 3 seconds
      setTimeout(() => {
        setUploadStatus({status: 'idle', message: ''});
      }, 3000);
    } catch (error) {
      console.error('File upload error:', error);
      setUploadStatus({status: 'error', message: `Upload failed: ${error instanceof Error ? error.message : 'Unknown error'}`});
    }
  };

  const handleChartClick = (params: unknown) => {
    // TODO: Implement chart interaction for visual query loop
    console.log('Chart clicked:', params);
  };

  const handleVisualQuerySubmit = () => {
    if (visualQuery.trim() && selectedArea) {
      // Send visual query to backend
      sendVisualQuery(visualQuery, { selection: selectedArea });
      
      // Reset query input
      setVisualQuery('');
      setShowQueryInput(false);
    }
  };

  const triggerFileInput = () => {
    fileInputRef.current?.click();
  };

  return (
    <div className="flex-1 flex flex-col bg-white">
      {/* Panel Header */}
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Visualizations</h2>
          <div className="flex items-center space-x-2">
            <button 
              className={`p-2 transition-colors ${isSelectionMode ? 'text-blue-600 bg-blue-100' : 'text-gray-400 hover:text-gray-600'}`}
              onClick={() => setShowQueryInput(!showQueryInput)}
              title="Visual Query"
              aria-label="Visual Query"
              disabled={!uploadedData}
            >
              <Square className="w-5 h-5" />
            </button>
            <button 
              className="p-2 text-gray-400 hover:text-gray-600 transition-colors"
              title="Bar Chart"
              aria-label="Bar Chart"
            >
              <BarChart3 className="w-5 h-5" />
            </button>
            <button 
              className="p-2 text-gray-400 hover:text-gray-600 transition-colors"
              title="Activity"
              aria-label="Activity"
            >
              <Activity className="w-5 h-5" />
            </button>
          </div>
        </div>
      </div>

      {/* Visual Query Input */}
      {showQueryInput && selectedArea && (
        <div className="px-6 py-4 border-b border-gray-200 bg-blue-50">
          <div className="flex items-center space-x-2">
            <input
              type="text"
              value={visualQuery}
              onChange={(e) => setVisualQuery(e.target.value)}
              placeholder="Ask a question about your selection..."
              className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  handleVisualQuerySubmit();
                }
              }}
              aria-label="Visual query input"
            />
            <button
              onClick={handleVisualQuerySubmit}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              aria-label="Submit visual query"
            >
              Ask
            </button>
            <button
              onClick={() => {
                setShowQueryInput(false);
                setVisualQuery('');
              }}
              className="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
              aria-label="Cancel visual query"
            >
              Cancel
            </button>
          </div>
          <p className="text-xs text-gray-500 mt-2">
            Selection area
          </p>
        </div>
      )}

      {/* Content Area */}
      <div 
        ref={chartContainerRef}
        className="flex-1 p-6"
      >
        {activeChart ? (
          // Show active chart
          <ChartRenderer
            chartData={activeChart}
            title="AI Generated Visualization"
            onChartClick={handleChartClick}
            className="h-full"
          />
        ) : uploadedData ? (
          // Show data uploaded state
          <div className="flex items-center justify-center h-full">
            <div className="text-center max-w-md">
              <CheckCircle className="w-16 h-16 text-green-500 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                Data Ready for Analysis
              </h3>
              <p className="text-gray-500 text-sm mb-4">
                Your data has been processed successfully. Ask me to create visualizations by typing requests like:
              </p>
              <div className="bg-gray-50 rounded-lg p-3 text-left text-sm mb-4">
                <p className="text-gray-700 mb-1">• &quot;Show me a bar chart of sales by month&quot;</p>
                <p className="text-gray-700 mb-1">• &quot;Create a scatter plot of price vs quantity&quot;</p>
                <p className="text-gray-700">• &quot;Display the distribution of categories&quot;</p>
              </div>
              <div className="text-left bg-blue-50 p-3 rounded-lg">
                <p className="text-sm font-medium text-blue-800">Uploaded File:</p>
                <p className="text-sm text-blue-700">{uploadedData.fileName}</p>
                <p className="text-xs text-blue-600 mt-1">
                  {uploadedData.rowCount} rows, {uploadedData.columnCount} columns
                </p>
              </div>
            </div>
          </div>
        ) : (
          // Show upload prompt
          <div
            className={`flex items-center justify-center h-full border-2 border-dashed rounded-lg transition-colors ${
              dragOver
                ? 'border-blue-400 bg-blue-50'
                : 'border-gray-300 hover:border-gray-400'
            }`}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
          >
            <div className="text-center max-w-md">
              <Upload className={`w-16 h-16 mx-auto mb-4 ${
                dragOver ? 'text-blue-500' : 'text-gray-400'
              }`} />
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                {dragOver ? 'Drop your file here' : 'Upload Your Dataset'}
              </h3>
              <p className="text-gray-500 text-sm mb-4">
                Upload a CSV file to start analyzing your data with AI-powered insights and visualizations.
              </p>
              
              {/* Upload status */}
              {uploadStatus.status !== 'idle' && (
                <div className={`mb-4 p-2 rounded text-sm flex items-center justify-center ${
                  uploadStatus.status === 'error' 
                    ? 'bg-red-100 text-red-700' 
                    : uploadStatus.status === 'success'
                    ? 'bg-green-100 text-green-700'
                    : 'bg-blue-100 text-blue-700'
                }`}>
                  {uploadStatus.status === 'uploading' && (
                    <>
                      <Activity className="w-4 h-4 mr-2 animate-spin" />
                      {uploadStatus.message}
                    </>
                  )}
                  {uploadStatus.status === 'success' && (
                    <>
                      <CheckCircle className="w-4 h-4 mr-2" />
                      {uploadStatus.message}
                    </>
                  )}
                  {uploadStatus.status === 'error' && (
                    <>
                      <AlertCircle className="w-4 h-4 mr-2" />
                      {uploadStatus.message}
                    </>
                  )}
                </div>
              )}
              
              <button 
                className="px-6 py-3 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 transition-colors"
                onClick={triggerFileInput}
              >
                Choose File
              </button>
              <input
                ref={fileInputRef}
                id="file-input"
                type="file"
                accept=".csv"
                className="hidden"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) handleFileUpload(file);
                }}
                aria-label="File input"
              />
              <p className="text-xs text-gray-400 mt-2">
                Supports CSV files up to 50MB
              </p>
            </div>
          </div>
        )}

        {/* Chart History */}
        {chartHistory.length > 0 && (
          <div className="mt-6">
            <h4 className="text-sm font-medium text-gray-900 mb-3">Chart History</h4>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {chartHistory.slice().reverse().slice(0, 6).map((chart, index) => (
                <div key={index} className="relative">
                  <ChartRenderer
                    chartData={chart}
                    className="h-32 cursor-pointer hover:ring-2 hover:ring-blue-500"
                    onChartClick={() => {
                      const { setActiveChart } = useChatStore.getState();
                      setActiveChart(chart);
                    }}
                  />
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Status Bar */}
      <div className="px-6 py-3 bg-gray-50 border-t border-gray-200">
        <div className="flex items-center justify-between text-sm text-gray-500">
          <span>
            {activeChart 
              ? 'Chart displayed' 
              : uploadedData 
              ? 'Data ready - ask for visualizations' 
              : 'Ready for data upload'
            }
          </span>
          <div className="flex items-center space-x-4">
            {chartHistory.length > 0 && (
              <span>{chartHistory.length} chart{chartHistory.length !== 1 ? 's' : ''} created</span>
            )}
            <div className="flex items-center space-x-1">
              <div className="w-2 h-2 bg-green-400 rounded-full"></div>
              <span>Connected</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
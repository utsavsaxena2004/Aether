'use client';

import React, { useEffect, useRef, useState } from 'react';
import * as echarts from 'echarts';
import { AlertCircle, Maximize2, Download, RefreshCw, Square, X } from 'lucide-react';
import { VisualSelectionOverlay } from './VisualSelectionOverlay';
import { useChatStore } from '@/store/chatStore';
import { ChartSpec } from '@/types/websocket';

interface ChartRendererProps {
    chartData: ChartSpec | string;
    title?: string;
    onChartClick?: (params: echarts.EChartsCoreOption) => void;
    className?: string;
}

export function ChartRenderer({ chartData, title, onChartClick, className = '' }: ChartRendererProps) {
    const chartRef = useRef<HTMLDivElement>(null);
    const chartInstanceRef = useRef<echarts.ECharts | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [showFullscreen, setShowFullscreen] = useState(false);
    const { isSelectionMode, setIsSelectionMode, setSelectedArea } = useChatStore();

    useEffect(() => {
        if (!chartRef.current || !chartData) {
            setIsLoading(false);
            return;
        }

        try {
            setError(null);
            setIsLoading(true);

            // Clean up existing chart
            if (chartInstanceRef.current) {
                chartInstanceRef.current.dispose();
            }

            // Parse chart data if it's a string
            let parsedData = chartData;
            if (typeof chartData === 'string') {
                try {
                    parsedData = JSON.parse(chartData);
                } catch (parseError) {
                    throw new Error('Invalid JSON format in chart data');
                }
            }

            // Validate chart data structure
            if (!parsedData || typeof parsedData !== 'object') {
                throw new Error('Invalid chart data format');
            }

            // Create new chart instance
            const chart = echarts.init(chartRef.current, 'default', {
                renderer: 'canvas',
                useDirtyRect: false,
            });

            // Set default responsive options
            const defaultOptions = {
                animation: true,
                animationDuration: 750,
                responsive: true,
                maintainAspectRatio: false,
            };

            // Merge with provided options
            const finalOptions = {
                ...defaultOptions,
                ...parsedData,
            };

            // Apply chart configuration
            chart.setOption(finalOptions as echarts.EChartsCoreOption, true);

            // Add click handler if provided
            if (onChartClick) {
                chart.on('click', onChartClick);
            }

            // Handle window resize
            const handleResize = () => {
                chart.resize();
            };
            window.addEventListener('resize', handleResize);

            chartInstanceRef.current = chart;
            setIsLoading(false);

            return () => {
                window.removeEventListener('resize', handleResize);
                if (chartInstanceRef.current) {
                    chartInstanceRef.current.dispose();
                    chartInstanceRef.current = null;
                }
            };
        } catch (err) {
            console.error('Chart rendering error:', err);
            setError(err instanceof Error ? err.message : 'Failed to render chart');
            setIsLoading(false);
        }
    }, [chartData, onChartClick]);

    const handleRefresh = () => {
        if (chartInstanceRef.current) {
            chartInstanceRef.current.resize();
        }
    };

    const handleDownload = () => {
        if (chartInstanceRef.current) {
            try {
                const url = chartInstanceRef.current.getDataURL({
                    pixelRatio: 2,
                    backgroundColor: '#fff',
                });
                const link = document.createElement('a');
                link.download = `chart-${Date.now()}.png`;
                link.href = url;
                link.click();
            } catch (err) {
                console.error('Failed to download chart:', err);
                setError('Failed to download chart image');
            }
        }
    };

    const handleMaximize = () => {
        setShowFullscreen(true);
    };

    const handleCloseFullscreen = () => {
        setShowFullscreen(false);
    };

    const toggleSelectionMode = () => {
        setIsSelectionMode(!isSelectionMode);
    };

    const handleSelectionComplete = (selection: { startX: number; startY: number; endX: number; endY: number; width: number; height: number }) => {
        // Store the selection
        setSelectedArea({
            startX: selection.startX,
            startY: selection.startY,
            endX: selection.endX,
            endY: selection.endY
        });

        // Exit selection mode
        setIsSelectionMode(false);

        // TODO: Send visual query to backend
        console.log('Visual selection completed:', selection);
    };

    if (error) {
        return (
            <div className={`flex items-center justify-center p-8 bg-red-50 border border-red-200 rounded-lg ${className}`}>
                <div className="text-center">
                    <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-3" />
                    <h3 className="text-lg font-medium text-red-900 mb-2">Chart Error</h3>
                    <p className="text-red-700 text-sm mb-4">{error}</p>
                    <div className="flex space-x-2 justify-center">
                        <button
                            onClick={handleRefresh}
                            className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
                        >
                            Try Again
                        </button>
                        <button
                            onClick={() => setError(null)}
                            className="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
                        >
                            Dismiss
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <>
            <div className={`relative bg-white rounded-lg border border-gray-200 ${className}`}>
                {/* Chart Header */}
                {title && (
                    <div className="px-4 py-3 border-b border-gray-200 flex items-center justify-between">
                        <h3 className="text-lg font-medium text-gray-900">{title}</h3>
                        <div className="flex items-center space-x-2">
                            <button
                                onClick={toggleSelectionMode}
                                className={`p-1 transition-colors ${isSelectionMode ? 'text-blue-600 bg-blue-100' : 'text-gray-400 hover:text-gray-600'}`}
                                title={isSelectionMode ? "Exit Selection Mode" : "Enter Selection Mode"}
                                aria-label={isSelectionMode ? "Exit Selection Mode" : "Enter Selection Mode"}
                            >
                                <Square className="w-4 h-4" />
                            </button>
                            <button
                                onClick={handleRefresh}
                                className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
                                title="Refresh Chart"
                                aria-label="Refresh Chart"
                            >
                                <RefreshCw className="w-4 h-4" />
                            </button>
                            <button
                                onClick={handleDownload}
                                className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
                                title="Download Chart"
                                aria-label="Download Chart"
                            >
                                <Download className="w-4 h-4" />
                            </button>
                            <button
                                onClick={handleMaximize}
                                className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
                                title="Maximize Chart"
                                aria-label="Maximize Chart"
                            >
                                <Maximize2 className="w-4 h-4" />
                            </button>
                        </div>
                    </div>
                )}

                {/* Chart Container */}
                <div className="relative">
                    {isLoading && (
                        <div className="absolute inset-0 flex items-center justify-center bg-gray-50 bg-opacity-75 z-10">
                            <div className="flex items-center space-x-2 text-gray-600">
                                <RefreshCw className="w-5 h-5 animate-spin" />
                                <span>Rendering chart...</span>
                            </div>
                        </div>
                    )}
                    <div
                        ref={chartRef}
                        className="w-full h-96 relative"
                        style={{ minHeight: '400px' }}
                    >
                        {chartData && (
                            <VisualSelectionOverlay
                                chartRef={chartRef as React.RefObject<HTMLDivElement>}
                                onSelectionComplete={handleSelectionComplete}
                            />
                        )}
                    </div>
                </div>
            </div>

            {/* Fullscreen Modal */}
            {showFullscreen && (
                <div className="fixed inset-0 bg-black bg-opacity-75 z-50 flex items-center justify-center p-4">
                    <div className="relative w-full h-full max-w-6xl max-h-[90vh]">
                        <button
                            onClick={handleCloseFullscreen}
                            className="absolute top-4 right-4 p-2 bg-white rounded-full shadow-lg z-10 hover:bg-gray-100"
                            aria-label="Close fullscreen"
                        >
                            <X className="w-5 h-5" />
                        </button>
                        <div className="w-full h-full bg-white rounded-lg overflow-hidden">
                            <ChartRenderer
                                chartData={chartData}
                                title={title}
                                onChartClick={onChartClick}
                                className="h-full"
                            />
                        </div>
                    </div>
                </div>
            )}
        </>
    );
}
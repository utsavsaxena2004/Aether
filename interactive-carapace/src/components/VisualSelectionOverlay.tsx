'use client';

import React, { useState, useEffect } from 'react';
import { useChatStore } from '@/store/chatStore';

interface SelectionArea {
  startX: number;
  startY: number;
  endX: number;
  endY: number;
  width: number;
  height: number;
}

interface VisualSelectionOverlayProps {
  chartRef: React.RefObject<HTMLDivElement>;
  onSelectionComplete: (selection: SelectionArea) => void;
}

export function VisualSelectionOverlay({ 
  chartRef, 
  onSelectionComplete 
}: VisualSelectionOverlayProps) {
  const [isSelecting, setIsSelecting] = useState(false);
  const [selection, setSelection] = useState<SelectionArea | null>(null);
  const isSelectionMode = useChatStore((state) => state.isSelectionMode);

  useEffect(() => {
    if (!chartRef.current || !isSelectionMode) return;

    const chartElement = chartRef.current;

    const handleMouseDown = (e: MouseEvent) => {
      if (!isSelectionMode) return;
      
      const rect = chartElement.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;
      
      setIsSelecting(true);
      setSelection({
        startX: x,
        startY: y,
        endX: x,
        endY: y,
        width: 0,
        height: 0
      });
    };

    const handleMouseMove = (e: MouseEvent) => {
      if (!isSelecting || !selection || !isSelectionMode) return;
      
      const rect = chartElement.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;
      
      const width = Math.abs(x - selection.startX);
      const height = Math.abs(y - selection.startY);
      
      setSelection({
        ...selection,
        endX: x,
        endY: y,
        width,
        height
      });
    };

    const handleMouseUp = () => {
      if (!isSelecting || !selection || !isSelectionMode) return;
      
      // Only send selection if it's large enough to be meaningful
      if (selection.width > 10 && selection.height > 10) {
        onSelectionComplete(selection);
      }
      
      setIsSelecting(false);
      setSelection(null);
    };

    // Attach event listeners
    chartElement.addEventListener('mousedown', handleMouseDown);
    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);

    return () => {
      chartElement.removeEventListener('mousedown', handleMouseDown);
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isSelecting, selection, isSelectionMode, chartRef, onSelectionComplete]);

  if (!isSelectionMode || !selection) return null;

  const left = Math.min(selection.startX, selection.endX);
  const top = Math.min(selection.startY, selection.endY);
  const width = selection.width;
  const height = selection.height;

  return (
    <div className="absolute inset-0 bg-transparent cursor-crosshair" style={{ zIndex: 1000 }}>
      {selection && (
        <div
          className="absolute border-2 border-blue-500 bg-blue-200 bg-opacity-30"
          style={{
            left: `${left}px`,
            top: `${top}px`,
            width: `${width}px`,
            height: `${height}px`,
          }}
        >
          <div className="absolute -top-8 left-0 text-xs text-blue-700 bg-blue-100 px-2 py-1 rounded">
            Selection: {Math.round(width)} × {Math.round(height)} px
          </div>
        </div>
      )}
    </div>
  );
}
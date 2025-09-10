// Message types for WebSocket communication
export interface Message {
  id: string;
  type:
    | "chat"
    | "chart_spec"
    | "error"
    | "system"
    | "user_connect"
    | "user_disconnect"
    | "visual_query";
  content?: string;
  data?: unknown;
  author: "user" | "ai" | "system";
  timestamp: Date;
  sessionId: string;
  isStreaming?: boolean;
}

// WebSocket connection states
export type ConnectionState =
  | "connecting"
  | "connected"
  | "disconnected"
  | "error";

// WebSocket hook return type
export interface UseWebSocketReturn {
  connectionState: ConnectionState;
  sendMessage: (
    message: Omit<Message, "id" | "timestamp" | "sessionId">
  ) => void;
  sendVisualQuery: (content: string, data: unknown) => void;
  messages: Message[];
  isConnected: boolean;
  sessionId: string | null;
  error: string | null;
}

// Chart data types (for future use)
export interface ChartData {
  type: string;
  data: unknown;
  options?: unknown;
}

// File upload types
export interface FileUpload {
  file: File;
  name: string;
  size: number;
  type: string;
}

// Visual selection area
export interface SelectionArea {
  startX: number;
  startY: number;
  endX: number;
  endY: number;
  width: number;
  height: number;
}

// Chart specification interface
export interface ChartSpec {
  title?: string;
  xAxis?: {
    type?: string;
    name?: string;
    data?: unknown[];
  };
  yAxis?: {
    type?: string;
    name?: string;
  };
  series?: Array<{
    name?: string;
    type?: string;
    data?: unknown[];
    [key: string]: unknown;
  }>;
  [key: string]: unknown;
}

// Uploaded data interface
export interface UploadedData {
  fileName: string;
  fileSize?: number;
  fileType?: string;
  rowCount: number;
  columnCount: number;
  columns?: string[];
  sampleData?: unknown[];
  [key: string]: unknown;
}
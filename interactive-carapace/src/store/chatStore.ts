import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { Message, ConnectionState, ChartSpec, UploadedData } from '@/types/websocket';

interface ChatState {
  // Messages
  messages: Message[];
  addMessage: (message: Message) => void;
  clearMessages: () => void;
  updateMessage: (id: string, updates: Partial<Message>) => void;
  removeMessage: (id: string) => void;
  
  // Connection state
  connectionState: ConnectionState;
  setConnectionState: (state: ConnectionState) => void;
  
  // Session management
  sessionId: string | null;
  setSessionId: (id: string | null) => void;
  
  // Error handling
  error: string | null;
  setError: (error: string | null) => void;
  
  // UI state
  isTyping: boolean;
  setIsTyping: (typing: boolean) => void;
  
  // Visual selection mode
  isSelectionMode: boolean;
  setIsSelectionMode: (mode: boolean) => void;
  selectedArea: { startX: number; startY: number; endX: number; endY: number } | null;
  setSelectedArea: (area: { startX: number; startY: number; endX: number; endY: number } | null) => void;
  
  // Data state
  uploadedData: UploadedData | null;
  setUploadedData: (data: UploadedData | null) => void;
  
  // Chart state
  activeChart: ChartSpec | null;
  chartHistory: ChartSpec[];
  setActiveChart: (chart: ChartSpec | null) => void;
  addChart: (chart: ChartSpec) => void;
  removeChart: (index: number) => void;
  clearCharts: () => void;
}

export const useChatStore = create<ChatState>()(
  devtools(
    persist(
      (set) => ({
        // Messages
        messages: [],
        addMessage: (message: Message) =>
          set((state) => ({
            messages: [...state.messages, message],
          })),
        clearMessages: () => set({ messages: [] }),
        updateMessage: (id: string, updates: Partial<Message>) =>
          set((state) => ({
            messages: state.messages.map((msg) =>
              msg.id === id ? { ...msg, ...updates } : msg
            ),
          })),
        removeMessage: (id: string) =>
          set((state) => ({
            messages: state.messages.filter((msg) => msg.id !== id),
          })),

        // Connection state
        connectionState: 'disconnected',
        setConnectionState: (connectionState: ConnectionState) =>
          set({ connectionState }),

        // Session management
        sessionId: null,
        setSessionId: (sessionId: string | null) => set({ sessionId }),

        // Error handling
        error: null,
        setError: (error: string | null) => set({ error }),

        // UI state
        isTyping: false,
        setIsTyping: (isTyping: boolean) => set({ isTyping }),
        
        // Visual selection mode
        isSelectionMode: false,
        setIsSelectionMode: (isSelectionMode: boolean) => set({ isSelectionMode }),
        selectedArea: null,
        setSelectedArea: (selectedArea: { startX: number; startY: number; endX: number; endY: number } | null) => 
          set({ selectedArea }),

        // Data state
        uploadedData: null,
        setUploadedData: (uploadedData: UploadedData | null) => set({ uploadedData }),

        // Chart state
        activeChart: null,
        chartHistory: [],
        setActiveChart: (activeChart: ChartSpec | null) => set({ activeChart }),
        addChart: (chart: ChartSpec) =>
          set((state) => ({
            chartHistory: [...state.chartHistory, chart],
            activeChart: chart,
          })),
        removeChart: (index: number) =>
          set((state) => {
            const newHistory = [...state.chartHistory];
            newHistory.splice(index, 1);
            return {
              chartHistory: newHistory,
              activeChart: newHistory.length > 0 ? newHistory[newHistory.length - 1] : null,
            };
          }),
        clearCharts: () => set({ chartHistory: [], activeChart: null }),
      }),
      {
        name: 'aether-chat-storage',
        partialize: (state) => ({
          messages: state.messages,
          sessionId: state.sessionId,
          uploadedData: state.uploadedData,
          chartHistory: state.chartHistory,
        }),
      }
    ),
    {
      name: 'chat-store',
    }
  )
);

// Selectors for better performance
export const useMessages = () => useChatStore((state) => state.messages);
export const useConnectionState = () => useChatStore((state) => state.connectionState);
export const useSessionId = () => useChatStore((state) => state.sessionId);
export const useError = () => useChatStore((state) => state.error);
export const useIsTyping = () => useChatStore((state) => state.isTyping);
export const useIsSelectionMode = () => useChatStore((state) => state.isSelectionMode);
export const useSelectedArea = () => useChatStore((state) => state.selectedArea);
export const useUploadedData = () => useChatStore((state) => state.uploadedData);
export const useActiveChart = () => useChatStore((state) => state.activeChart);
export const useChartHistory = () => useChatStore((state) => state.chartHistory);
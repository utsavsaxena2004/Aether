'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { UseWebSocketReturn, Message, ChartSpec } from '@/types/websocket';
import { useChatStore } from '@/store/chatStore';

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws';
const RECONNECT_INTERVAL = 3000;
const MAX_RECONNECT_ATTEMPTS = 5;

export function useWebSocket(): UseWebSocketReturn {
  const [reconnectAttempts, setReconnectAttempts] = useState(0);
  
  // Zustand store actions and state
  const {
    messages,
    addMessage,
    updateMessage,
    connectionState,
    setConnectionState,
    sessionId,
    setSessionId,
    error,
    setError,
    setIsTyping,
  } = useChatStore();
  
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const isIntentionalClose = useRef(false);
  const streamingMessagesRef = useRef<Record<string, Message>>({});

  const connect = useCallback(() => {
    // Close existing connection if open
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.close();
      wsRef.current = null;
    }

    setConnectionState('connecting');
    setError(null);

    try {
      console.log(`🔌 Attempting to connect to WebSocket at ${WS_URL}`);
      const ws = new WebSocket(WS_URL);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('🔌 WebSocket connected successfully');
        setConnectionState('connected');
        setReconnectAttempts(0);
        setError(null);
      };

      ws.onmessage = (event) => {
        try {
          const message: Message = JSON.parse(event.data);
          
          // Set session ID from first message if not already set
          if (!sessionId && message.sessionId) {
            setSessionId(message.sessionId);
          }
          
          // Convert timestamp string to Date object
          if (typeof message.timestamp === 'string') {
            message.timestamp = new Date(message.timestamp);
          }
          
          // Handle streaming messages
          if (message.isStreaming) {
            // Update streaming message in progress
            streamingMessagesRef.current[message.id] = message;
            setIsTyping(true);
            
            // Update or add the message in the store
            const existingMessage = messages.find(msg => msg.id === message.id);
            if (existingMessage) {
              updateMessage(message.id, { content: message.content });
            } else {
              addMessage(message);
            }
            return;
          } else {
            // Final message of a stream or regular message
            // Remove from streaming messages
            delete streamingMessagesRef.current[message.id];
            
            // If no more streaming messages, stop typing indicator
            if (Object.keys(streamingMessagesRef.current).length === 0) {
              setIsTyping(false);
            }
          }
          
          // Handle different message types
          if (message.type === 'chart_spec') {
            // Handle chart specification message
            const { addChart } = useChatStore.getState();
            // Type assertion since we trust the backend to send the correct data structure
            addChart(message.data as ChartSpec);
          }
          
          addMessage(message);
          console.log('📨 Received message:', message);
        } catch (err) {
          console.error('❌ Failed to parse WebSocket message:', err);
          setError('Failed to parse incoming message');
        }
      };

      ws.onclose = (event) => {
        console.log('🔌 WebSocket disconnected:', event.code, event.reason);
        wsRef.current = null;
        setConnectionState('disconnected');
        
        // Stop typing indicator when disconnected
        setIsTyping(false);
        streamingMessagesRef.current = {};
        
        // Only attempt to reconnect if it wasn't an intentional close
        if (!isIntentionalClose.current && reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
          console.log(`🔁 Attempting to reconnect (${reconnectAttempts + 1}/${MAX_RECONNECT_ATTEMPTS})...`);
          const timeout = setTimeout(() => {
            setReconnectAttempts(prev => prev + 1);
            connect();
          }, RECONNECT_INTERVAL);
          reconnectTimeoutRef.current = timeout;
        } else if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
          const errorMsg = 'Failed to connect after multiple attempts. Please check if the backend is running.';
          console.error(errorMsg);
          setError(errorMsg);
          setConnectionState('error');
        }
      };

      ws.onerror = (event) => {
        console.error('❌ WebSocket error:', event);
        const errorMsg = 'WebSocket connection error. Please check if the backend is running.';
        setError(errorMsg);
        setConnectionState('error');
      };
    } catch (err) {
      console.error('❌ Failed to create WebSocket connection:', err);
      const errorMsg = 'Failed to create WebSocket connection. Please check if the backend is running.';
      setError(errorMsg);
      setConnectionState('error');
    }
  }, [sessionId, reconnectAttempts, messages, addMessage, updateMessage, setConnectionState, setError, setIsTyping, setSessionId]);

  const disconnect = useCallback(() => {
    isIntentionalClose.current = true;
    
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    
    if (wsRef.current) {
      wsRef.current.close(1000, 'Client initiated close');
      wsRef.current = null;
    }
    
    setConnectionState('disconnected');
    setIsTyping(false);
    streamingMessagesRef.current = {};
  }, [setConnectionState, setIsTyping]);

  const sendMessage = useCallback((message: Omit<Message, 'id' | 'timestamp' | 'sessionId'>) => {
    if (wsRef.current?.readyState !== WebSocket.OPEN) {
      console.warn('⚠️ WebSocket not connected, cannot send message');
      setError('Not connected to server');
      return;
    }

    const fullMessage: Message = {
      ...message,
      id: generateMessageId(),
      timestamp: new Date(),
      sessionId: sessionId || '',
    };

    try {
      wsRef.current.send(JSON.stringify(fullMessage));
      console.log('📤 Sent message:', fullMessage);
      
      // Add the message to store for immediate UI feedback
      if (message.author === 'user') {
        addMessage(fullMessage);
      }
    } catch (err) {
      console.error('❌ Failed to send message:', err);
      setError('Failed to send message');
    }
  }, [sessionId, addMessage, setError]);

  // Function to send visual query messages
  const sendVisualQuery = useCallback((content: string, data: unknown) => {
    if (wsRef.current?.readyState !== WebSocket.OPEN) {
      console.warn('⚠️ WebSocket not connected, cannot send visual query');
      setError('Not connected to server');
      return;
    }

    const visualQueryMessage: Omit<Message, 'id' | 'timestamp' | 'sessionId'> = {
      type: 'visual_query',
      content: content,
      data: data,
      author: 'user',
    };

    sendMessage(visualQueryMessage);
  }, [sendMessage, setError]);

  // Auto-connect on mount
  useEffect(() => {
    isIntentionalClose.current = false;
    connect();

    // Cleanup on unmount
    return () => {
      isIntentionalClose.current = true;
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    connectionState,
    sendMessage,
    sendVisualQuery,
    messages,
    isConnected: connectionState === 'connected',
    sessionId,
    error,
  };
}

// Helper function to generate unique message IDs
function generateMessageId(): string {
  return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}
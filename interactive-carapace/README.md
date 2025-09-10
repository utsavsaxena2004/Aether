# Interactive Carapace - Aether Frontend

This is the frontend component of the Aether project, built with Next.js 14+ and TypeScript. It provides a responsive, interactive interface for conversational data analysis with AI-powered visualizations.

## Features

- **Real-time Chat Interface**: Communicate with the AI assistant via WebSocket
- **Interactive Data Visualizations**: Dynamic charts powered by ECharts
- **Visual Query Loop**: Select areas on charts and ask follow-up questions
- **File Upload**: Upload CSV files for analysis
- **Session Management**: Persistent sessions using Redis
- **Responsive Design**: Works on desktop and mobile devices

## Tech Stack

- [Next.js 14+](https://nextjs.org/) - React framework with App Router
- [TypeScript](https://www.typescriptlang.org/) - Type-safe JavaScript
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
- [ECharts](https://echarts.apache.org/) - Data visualization library
- [Zustand](https://github.com/pmndrs/zustand) - State management
- [Lucide React](https://lucide.dev/) - Icon library

## Getting Started

### Prerequisites

- Node.js >=18.0.0
- pnpm >=8.0.0

### Development

First, install dependencies:

```bash
pnpm install
```

Then, run the development server:

```bash
pnpm dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

### Environment Variables

Create a `.env.local` file based on `.env.local.example`:

```bash
cp .env.local.example .env.local
```

Configure the following variables:

```env
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Build for Production

```bash
pnpm build
```

### Start Production Server

```bash
pnpm start
```

## Project Structure

```
src/
├── app/                 # Next.js app router pages
│   ├── layout.tsx       # Root layout
│   └── page.tsx         # Home page
├── components/          # React components
│   ├── ChartRenderer.tsx        # ECharts renderer
│   ├── ChatPanel.tsx            # Chat interface
│   ├── Dashboard.tsx            # Main dashboard layout
│   ├── Header.tsx               # Application header
│   ├── VisualSelectionOverlay.tsx # Visual selection tool
│   └── VisualizationPanel.tsx   # Visualization area
├── hooks/               # Custom React hooks
│   └── useWebSocket.ts  # WebSocket connection hook
├── store/               # Zustand store
│   └── chatStore.ts     # Chat state management
└── types/               # TypeScript types
    └── websocket.ts     # WebSocket message types
```

## Key Components

### Dashboard
The main layout component that includes the sidebar navigation and header.

### VisualizationPanel
The main area for data visualization, including file upload functionality and chart display.

### ChatPanel
The chat interface for communicating with the AI assistant.

### ChartRenderer
Component responsible for rendering ECharts visualizations.

### WebSocket Hook
Custom hook for managing WebSocket connections and message handling.

## State Management

The application uses Zustand for state management, with a centralized store for:
- Chat messages
- Connection status
- Session information
- Error states
- UI states (typing indicators, selection mode)
- Data and chart history

## WebSocket Communication

The frontend communicates with the backend via WebSocket for real-time interaction:
- Chat messages
- Chart specifications
- Visual queries
- System messages

## Styling

The application uses Tailwind CSS for styling with a clean, modern design. All components are responsive and accessible.

## Learn More

To learn more about the technologies used:

- [Next.js Documentation](https://nextjs.org/docs)
- [TypeScript Documentation](https://www.typescriptlang.org/docs/)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [ECharts Documentation](https://echarts.apache.org/en/documents.html)

## Deployment

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new) from the creators of Next.js.

Check out the [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
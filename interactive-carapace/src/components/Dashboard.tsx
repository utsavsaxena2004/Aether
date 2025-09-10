'use client';

import React, { useState, useEffect } from 'react';
import { 
  BarChart3, 
  PieChart, 
  LineChart, 
  Database, 
  Upload, 
  Settings, 
  Bell, 
  User, 
  Menu, 
  X, 
  AlertCircle,
  CheckCircle,
  Wifi,
  WifiOff,
  Info
} from 'lucide-react';
import { VisualizationPanel } from './VisualizationPanel';
import { ChatPanel } from './ChatPanel';
import { useConnectionState } from '@/store/chatStore';

export function Dashboard() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [activeTab, setActiveTab] = useState('dashboard');
  const connectionState = useConnectionState();

  const menuItems = [
    { id: 'dashboard', label: 'Dashboard', icon: BarChart3 },
    { id: 'visualizations', label: 'Visualizations', icon: PieChart },
    { id: 'analytics', label: 'Analytics', icon: LineChart },
    { id: 'data', label: 'Data Sources', icon: Database },
    { id: 'uploads', label: 'Uploads', icon: Upload },
    { id: 'settings', label: 'Settings', icon: Settings },
  ];

  // Close sidebar when resizing to larger screens
  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth >= 1024) {
        setSidebarOpen(false);
      }
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const getConnectionStatusIcon = () => {
    switch (connectionState) {
      case 'connected': return <Wifi className="w-4 h-4 text-green-500" />;
      case 'error': return <WifiOff className="w-4 h-4 text-red-500" />;
      default: return <AlertCircle className="w-4 h-4 text-yellow-500" />;
    }
  };

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div 
          className="fixed inset-0 z-40 bg-black bg-opacity-50 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <div 
        className={`fixed inset-y-0 left-0 z-50 w-64 bg-white shadow-lg transform transition-transform duration-300 ease-in-out lg:translate-x-0 lg:static lg:inset-0 ${
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <div className="flex items-center justify-between h-16 px-4 border-b border-gray-200">
          <div className="flex items-center">
            <BarChart3 className="w-8 h-8 text-blue-600" />
            <span className="ml-2 text-xl font-bold text-gray-800">Aether</span>
          </div>
          <button 
            className="lg:hidden text-gray-500 hover:text-gray-700"
            onClick={() => setSidebarOpen(false)}
          >
            <X className="w-6 h-6" />
          </button>
        </div>
        
        {/* Connection Status */}
        <div className="px-4 py-3 border-b border-gray-200 bg-gray-50">
          <div className="flex items-center text-sm">
            {getConnectionStatusIcon()}
            <span className="ml-2 text-gray-600">
              {connectionState === 'connected' ? 'Connected' : 
               connectionState === 'error' ? 'Connection Error' : 'Connecting...'}
            </span>
          </div>
        </div>
        
        <nav className="mt-6 px-2">
          {menuItems.map((item) => {
            const Icon = item.icon;
            return (
              <button
                key={item.id}
                className={`flex items-center w-full px-4 py-3 text-sm font-medium rounded-lg transition-colors ${
                  activeTab === item.id
                    ? 'bg-blue-50 text-blue-700 border-r-2 border-blue-700'
                    : 'text-gray-700 hover:bg-gray-100'
                }`}
                onClick={() => {
                  setActiveTab(item.id);
                  setSidebarOpen(false);
                }}
              >
                <Icon className="w-5 h-5 mr-3" />
                {item.label}
              </button>
            );
          })}
        </nav>
        
        {/* Sidebar Footer */}
        <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-gray-200">
          <div className="flex items-center">
            <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center">
              <User className="w-5 h-5 text-white" />
            </div>
            <div className="ml-3">
              <p className="text-sm font-medium text-gray-900">User</p>
              <p className="text-xs text-gray-500">Free Plan</p>
            </div>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="bg-white shadow-sm">
          <div className="flex items-center justify-between h-16 px-4">
            <div className="flex items-center">
              <button 
                className="lg:hidden text-gray-500 hover:text-gray-700 mr-4"
                onClick={() => setSidebarOpen(true)}
              >
                <Menu className="w-6 h-6" />
              </button>
              <h1 className="text-xl font-semibold text-gray-800 capitalize">
                {menuItems.find(item => item.id === activeTab)?.label || 'Dashboard'}
              </h1>
            </div>
            
            <div className="flex items-center space-x-4">
              <button className="p-2 text-gray-500 hover:text-gray-700 rounded-full hover:bg-gray-100 relative" aria-label="Notifications">
                <Bell className="w-5 h-5" />
                <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full"></span>
              </button>
              <button className="p-2 text-gray-500 hover:text-gray-700 rounded-full hover:bg-gray-100" aria-label="Settings">
                <Settings className="w-5 h-5" />
              </button>
              <div className="flex items-center">
                <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center">
                  <User className="w-5 h-5 text-white" />
                </div>
              </div>
            </div>
          </div>
        </header>

        {/* Content area */}
        <main className="flex-1 overflow-hidden flex flex-col">
          {activeTab === 'dashboard' || activeTab === 'visualizations' ? (
            <div className="flex flex-col md:flex-row flex-1 overflow-hidden">
              {/* Visualization Panel - 70% width on medium screens and up */}
              <div className="w-full md:w-7/10 lg:w-3/5 xl:w-2/3 h-1/2 md:h-full overflow-hidden">
                <VisualizationPanel />
              </div>
              
              {/* Chat Panel - 30% width on medium screens and up */}
              <div className="w-full md:w-3/10 lg:w-2/5 xl:w-1/3 h-1/2 md:h-full border-l border-gray-200">
                <ChatPanel />
              </div>
            </div>
          ) : activeTab === 'uploads' ? (
            <div className="p-6 flex-1 overflow-auto">
              <div className="bg-white rounded-lg shadow p-6 max-w-4xl mx-auto">
                <div className="flex items-center mb-6">
                  <Upload className="w-8 h-8 text-blue-600 mr-3" />
                  <h2 className="text-2xl font-bold text-gray-900">File Uploads</h2>
                </div>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-4">Upload New Data</h3>
                    <p className="text-gray-600 mb-4">
                      Upload CSV files to analyze your data with our AI-powered tools.
                    </p>
                    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
                      <div className="flex items-start">
                        <Info className="w-5 h-5 text-blue-500 mr-2 mt-0.5 flex-shrink-0" />
                        <div>
                          <p className="text-sm text-blue-800 font-medium">Supported formats</p>
                          <p className="text-sm text-blue-700">CSV files up to 50MB</p>
                        </div>
                      </div>
                    </div>
                    <button 
                      className="w-full py-2 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                      onClick={() => setActiveTab('visualizations')}
                    >
                      Go to Visualization Panel
                    </button>
                  </div>
                  
                  <div className="border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-4">Recent Uploads</h3>
                    <div className="text-center py-8">
                      <Database className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                      <p className="text-gray-500">No recent uploads</p>
                      <p className="text-sm text-gray-400 mt-1">Upload your first file to get started</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          ) : (
            <div className="p-6 flex-1 overflow-auto">
              <div className="bg-white rounded-lg shadow p-6 max-w-4xl mx-auto">
                <div className="flex items-center mb-6">
                  <div className="p-2 bg-gray-100 rounded-lg mr-3">
                    {React.createElement(menuItems.find(item => item.id === activeTab)?.icon || BarChart3, { className: "w-6 h-6 text-gray-600" })}
                  </div>
                  <h2 className="text-2xl font-bold text-gray-900 capitalize">
                    {menuItems.find(item => item.id === activeTab)?.label}
                  </h2>
                </div>
                
                <div className="text-center py-12">
                  <div className="mx-auto w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
                    <CheckCircle className="w-8 h-8 text-gray-400" />
                  </div>
                  <h3 className="text-lg font-medium text-gray-900 mb-2">Feature Coming Soon</h3>
                  <p className="text-gray-500 max-w-md mx-auto">
                    We're working on implementing this feature. Check back soon for updates.
                  </p>
                </div>
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
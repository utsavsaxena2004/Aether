'use client';

import React from 'react';
import { Dashboard } from '@/components/Dashboard';

export default function Home() {
  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <Dashboard />
    </div>
  );
}

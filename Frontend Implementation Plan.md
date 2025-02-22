# Frontend Implementation Plan

## 1. Project Setup
```bash
# Create React TypeScript project
cd /home/karl/Documents/Genaibootcamp/lang-portal
npx create-react-app frontend --template typescript
cd frontend

# Install dependencies
npm install @mui/material @emotion/react @emotion/styled
npm install @tanstack/react-query
npm install react-router-dom
npm install axios
npm install date-fns
```

## 2. Project Structure
```bash
frontend/
├── src/
│   ├── components/      # Reusable components
│   ├── pages/          # Page components
│   ├── api/            # API client
│   ├── hooks/          # Custom hooks
│   ├── types/          # TypeScript types
│   ├── utils/          # Utility functions
│   └── context/        # React context
```

## 3. Type Definitions
```typescript
// filepath: /home/karl/Documents/Genaibootcamp/lang-portal/frontend/src/types/index.ts
export interface Word {
  id: number;
  japanese: string;
  romaji: string;
  english: string;
  correctCount: number;
  wrongCount: number;
}

export interface StudySession {
  id: number;
  activityName: string;
  groupName: string;
  startTime: string;
  endTime: string;
  reviewItemCount: number;
}

export interface DashboardStats {
  lastStudySession: StudySession;
  studyProgress: {
    totalWords: number;
    studiedWords: number;
    masteryPercentage: number;
  };
  quickStats: {
    successRate: number;
    totalSessions: number;
    activeGroups: number;
    studyStreak: number;
  };
}
```

## 4. API Client
```typescript
// filepath: /home/karl/Documents/Genaibootcamp/lang-portal/frontend/src/api/client.ts
import axios from 'axios';

const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080/api'
});

export const fetchDashboardStats = async (): Promise<DashboardStats> => {
  const [lastSession, progress, stats] = await Promise.all([
    api.get('/dashboard/last_study_session'),
    api.get('/dashboard/study_progress'),
    api.get('/dashboard/quick_stats')
  ]);

  return {
    lastStudySession: lastSession.data,
    studyProgress: progress.data,
    quickStats: stats.data
  };
};
```

## 5. Dashboard Implementation
```typescript
// filepath: /home/karl/Documents/Genaibootcamp/lang-portal/frontend/src/pages/Dashboard.tsx
import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchDashboardStats } from '../api/client';
import {
  Card,
  CardContent,
  Typography,
  Grid,
  LinearProgress,
  Button
} from '@mui/material';

export const Dashboard: React.FC = () => {
  const { data, isLoading } = useQuery(['dashboardStats'], fetchDashboardStats);

  if (isLoading) return <LinearProgress />;

  return (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Card>
          <CardContent>
            <Typography variant="h5">Last Study Session</Typography>
            {/* Last session details */}
          </CardContent>
        </Card>
      </Grid>
      
      <Grid item xs={12} md={6}>
        <Card>
          <CardContent>
            <Typography variant="h5">Study Progress</Typography>
            <LinearProgress 
              variant="determinate" 
              value={data?.studyProgress.masteryPercentage || 0} 
            />
          </CardContent>
        </Card>
      </Grid>

      <Grid item xs={12} md={6}>
        <Card>
          <CardContent>
            <Typography variant="h5">Quick Stats</Typography>
            {/* Stats display */}
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  );
};
```

## 6. Router Setup
```typescript
// filepath: /home/karl/Documents/Genaibootcamp/lang-portal/frontend/src/App.tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const queryClient = new QueryClient();

export const App: React.FC = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/study_activities" element={<StudyActivities />} />
          <Route path="/words" element={<Words />} />
          <Route path="/groups" element={<Groups />} />
          <Route path="/settings" element={<Settings />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
};
```
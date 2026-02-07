import { createBrowserRouter } from 'react-router-dom';
import { Layout } from '@/components/Layout';
import AuthGuard from '@/components/AuthGuard';
import Login from '@/pages/Login';
import Dashboard from '@/pages/Dashboard';
import NodeList from '@/pages/NodeList';
import NodeDetail from '@/pages/NodeDetail';
import JobList from '@/pages/JobList';
import JobDetail from '@/pages/JobDetail';
import Settings from '@/pages/Settings';

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '/',
    element: <AuthGuard><Layout /></AuthGuard>,
    children: [
      {
        index: true,
        element: <Dashboard />,
      },
      {
        path: 'nodes',
        element: <NodeList />,
      },
      {
        path: 'nodes/:nodeId',
        element: <NodeDetail />,
      },
      {
        path: 'jobs',
        element: <JobList />,
      },
      {
        path: 'jobs/:jobId',
        element: <JobDetail />,
      },
      {
        path: 'settings',
        element: <Settings />,
      },
    ],
  },
]);

import { createBrowserRouter } from 'react-router-dom';
import { Layout } from '@/components/Layout';
import Dashboard from '@/pages/Dashboard';
import NodeList from '@/pages/NodeList';
import NodeDetail from '@/pages/NodeDetail';
import JobList from '@/pages/JobList';
import JobDetail from '@/pages/JobDetail';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
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
    ],
  },
]);

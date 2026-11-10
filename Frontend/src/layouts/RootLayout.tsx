import { Outlet } from 'react-router-dom';
import { RootTemplate } from '../components/templates';

export default function RootLayout() {
  return (
    <RootTemplate>
      <Outlet />
    </RootTemplate>
  );
}

import { useAtomValue } from 'jotai';
import { activeDownloadsState } from '../atoms/downloads';
import { listViewState } from '../atoms/settings';
import { ProcessStatus, RPCResult } from '../types';
import DownloadsGridView from './DownloadsGridView';
import DownloadsTableView from './DownloadsTableView';
import { Typography } from '@mui/material'; // Added for empty state
import { useI18n } from '../hooks/useI18n'; // Added for empty state text

const ActiveDownloadsDisplay: React.FC = () => {
  const allActiveDownloads = useAtomValue(activeDownloadsState);
  const tableView = useAtomValue(listViewState);
  const { i18n } = useI18n(); // For empty state text

  const activeDownloads = allActiveDownloads.filter(
    (d) =>
      d.progress.process_status === ProcessStatus.DOWNLOADING ||
      d.progress.process_status === ProcessStatus.LIVESTREAM
  );

  if (activeDownloads.length === 0) {
    return <Typography sx={{ p: 2 }}>{i18n.t('noActiveDownloads')}</Typography>;
  }

  if (tableView) {
    return <DownloadsTableView downloads={activeDownloads} />;
  }
  return <DownloadsGridView downloads={activeDownloads} />;
};

export default ActiveDownloadsDisplay;

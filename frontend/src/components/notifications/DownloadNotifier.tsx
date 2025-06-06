import React, { useEffect, useRef } from 'react';
import { useAtomValue } from 'jotai';
import { activeDownloadsState } from 'frontend/src/atoms/downloads';
import { useToast } from 'frontend/src/hooks/toast';
import { useI18n } from 'frontend/src/hooks/useI18n';
// Assuming RPCResult might be needed for the structure of download items,
// though not directly used in this component's logic based on the refined requirements.
// If download objects in activeDownloadsState don't conform to a part of RPCResult that includes progress,
// this import might be adjusted or removed if truly unused.
import { RPCResult } from 'frontend/src/types';

const DownloadNotifier: React.FC = () => {
  const { i18n } = useI18n();
  const { pushMessage } = useToast();
  const downloads = useAtomValue(activeDownloadsState);
  const notifiedDownloadsRef = useRef<Record<string, string>>({}); // Stores {id: 'completed'} or {id: 'errored'}

  useEffect(() => {
    const currentNotified = notifiedDownloadsRef.current;
    const activeIds = new Set(downloads.map(d => d.id));

    // Prune old notifications from ref if their IDs are no longer active
    for (const id in currentNotified) {
      if (!activeIds.has(id)) {
        delete currentNotified[id];
      }
    }

    downloads.forEach(download => {
      const id = download.id;
      // Ensure that download.progress is defined before accessing status, error
      const status = download.progress?.status;
      const title = download.name || download.title || id; // Use download.title as per RPCResult

      // Check for completed status (status code 2 or string 'completed')
      if (status === 2 || status === 'completed') {
        if (currentNotified[id] !== 'completed') {
          pushMessage(i18n.t('downloadCompletedNotif', { title }), 'success');
          currentNotified[id] = 'completed';
        }
      }
      // Check for errored status (status code 3 or string 'errored')
      else if (status === 3 || status === 'errored') {
        if (currentNotified[id] !== 'errored') {
          const errorMessage = download.progress?.error || i18n.t('unknownError');
          pushMessage(i18n.t('downloadErroredNotif', { title, error: errorMessage }), 'error');
          currentNotified[id] = 'errored';
        }
      }
    });
  }, [downloads, pushMessage, i18n]);

  return null; // Component does not render anything
};

export default DownloadNotifier;

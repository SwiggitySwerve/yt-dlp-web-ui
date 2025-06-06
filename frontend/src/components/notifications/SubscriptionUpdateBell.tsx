import React, { useState, useEffect, useCallback } from 'react';
import { IconButton, Badge, Tooltip } from '@mui/material';
import NotificationsIcon from '@mui/icons-material/Notifications';
import { useNavigate } from 'react-router-dom';
import { useAtomValue } from 'jotai';
import { serverURL } from '../../atoms/settings';
import { useI18n } from '../../hooks/useI18n';
import { ffetch } from '../../lib/httpClient'; // Using ffetch directly

const SubscriptionUpdateBell: React.FC = () => {
  const { i18n } = useI18n();
  const navigate = useNavigate();
  const apiBaseURL = useAtomValue(serverURL);
  const [unseenCount, setUnseenCount] = useState(0);

  const fetchUnseenCount = useCallback(async () => {
    if (!apiBaseURL) return; // Don't fetch if serverURL is not set

    try {
      const response = await ffetch<{ count: number }>(`${apiBaseURL}/api/subscriptions/updates/count`); // Assuming this endpoint
      if (response && typeof response.count === 'number') {
        setUnseenCount(response.count);
      } else {
        console.error('Invalid response structure from /api/subscriptions/updates/count:', response);
      }
    } catch (error) {
      console.error('Failed to fetch unseen subscription updates count:', error);
      // Optionally, set unseenCount to 0 or a specific error indicator if needed
    }
  }, [apiBaseURL]);

  useEffect(() => {
    fetchUnseenCount(); // Fetch on mount

    const intervalId = setInterval(fetchUnseenCount, 5 * 60 * 1000); // Fetch every 5 minutes

    return () => clearInterval(intervalId); // Cleanup timer on unmount
  }, [fetchUnseenCount]);

  const handleBellClick = () => {
    console.log('SubscriptionUpdateBell clicked. Navigating to /new_videos_updates (placeholder).');
    // For now, this is a placeholder. In a future step, it will navigate to a dedicated view or open a modal.
    // Example navigation (uncomment and adjust path when ready):
    // navigate('/new_videos_updates');
    // Or, if opening a modal, trigger modal state here.
    // For now, let's just reset count locally for demo purposes, actual reset should happen when user views updates
    // setUnseenCount(0);
    navigate('/new_videos_updates'); // Navigate to the new view
  };

  return (
    <Tooltip title={i18n.t('newVideoUpdatesTooltip', { count: unseenCount }) || "New video updates"}>
      <IconButton color="inherit" onClick={handleBellClick}>
        <Badge badgeContent={unseenCount} color="error" invisible={unseenCount === 0}>
          <NotificationsIcon />
        </Badge>
      </IconButton>
    </Tooltip>
  );
};

export default SubscriptionUpdateBell;

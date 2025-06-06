import React, { useState } from 'react';
import { Alert, AlertTitle, Typography, Box, Collapse, IconButton, Button } from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import { useI18n } from '../../hooks/useI18n';

// This interface should match the backend's ErrorResponse struct
interface BackendErrorResponse {
  message: string;
  detail?: string;
  code?: number;
}

interface ApiErrorDisplayProps {
  error: any; // The error object from useFetch, ffetch, or standard JS Error
  title?: string;
}

const ApiErrorDisplay: React.FC<ApiErrorDisplayProps> = ({ error, title }) => {
  const { i18n } = useI18n();
  const [expanded, setExpanded] = useState(false);

  let displayMessage: string = i18n.t('genericApiError') || "A generic error occurred. Please try again.";
  let errorDetail: string | null = null;

  if (error) {
    // Case 1: Error object from useFetch (error.data contains the parsed JSON response)
    if (error.data && typeof error.data === 'object' && error.data.message) {
      const backendError = error.data as BackendErrorResponse;
      displayMessage = backendError.message;
      errorDetail = backendError.detail || null;
    }
    // Case 2: Error object from ffetch (error.message might be a JSON string of the backend response)
    // or a standard JavaScript Error object.
    else if (typeof error.message === 'string') {
      displayMessage = error.message; // Default to the error's message
      try {
        // Attempt to parse if error.message is a JSON string
        const parsedJsonInMessage = JSON.parse(error.message) as BackendErrorResponse;
        if (parsedJsonInMessage && parsedJsonInMessage.message) {
            displayMessage = parsedJsonInMessage.message; // Override if JSON message is more specific
            errorDetail = parsedJsonInMessage.detail || error.stack || null;
        } else {
            // Parsed but not the expected structure, fall back to stack if available
             if (typeof error.stack === 'string') {
                errorDetail = error.stack;
            }
        }
      } catch (e) {
        // error.message was not a JSON string, use error.stack if available
        if (typeof error.stack === 'string') {
          errorDetail = error.stack;
        }
      }
    }
    // Case 3: Error is a string (e.g., a simple message thrown)
    else if (typeof error === 'string') {
      displayMessage = error;
    }
    // Case 4: Error is an object that directly matches BackendErrorResponse (e.g., already parsed by caller)
    else if (typeof error.detail === 'string' && typeof error.message === 'string') {
        const backendError = error as BackendErrorResponse;
        displayMessage = backendError.message;
        errorDetail = backendError.detail;
    }
    // Case 5: Fallback for other object types, try to stringify or get stack
    else if (typeof error === 'object') {
        displayMessage = JSON.stringify(error);
        if (typeof error.stack === 'string') {
            errorDetail = error.stack;
        }
    }
  }

  const alertTitle = title || (i18n.t('errorOccurredTitle') || "An Error Occurred");

  return (
    <Alert severity="error" sx={{ mt: 2, mb: 2, textAlign: 'left' }}>
      <AlertTitle>{alertTitle}</AlertTitle>
      <Typography variant="body2" component="div" sx={{ wordBreak: 'break-word' }}>{displayMessage}</Typography>
      {errorDetail && (
        <Box sx={{ mt: 1 }}>
          <Button
            size="small"
            onClick={() => setExpanded(!expanded)}
            startIcon={<ExpandMoreIcon sx={{ transform: expanded ? 'rotate(180deg)' : 'rotate(0deg)' }} />}
            aria-expanded={expanded}
            aria-label={expanded ? (i18n.t('hideErrorDetailsButton') || "Hide Details") : (i18n.t('showErrorDetailsButton') || "Details")}
          >
            {expanded ? (i18n.t('hideErrorDetailsButton') || "Hide Details") : (i18n.t('showErrorDetailsButton') || "Details")}
          </Button>
          <Collapse in={expanded} timeout="auto" unmountOnExit>
            <Typography variant="caption" component="pre" sx={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all', mt: 1, fontFamily: 'monospace' }}>
              {errorDetail}
            </Typography>
          </Collapse>
        </Box>
      )}
    </Alert>
  );
};

export default ApiErrorDisplay;

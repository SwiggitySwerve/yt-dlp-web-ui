import React from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Typography,
  Grid,
  Chip,
  Link,
  IconButton,
  Avatar, // Added Avatar for placeholder
  Divider, // Added Divider
} from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import MovieIcon from '@mui/icons-material/Movie'; // Placeholder for missing thumbnail
import { ArchiveEntry } from '../../types';
import { useI18n } from '../../hooks/useI18n';
import { formatDuration, formatDate } from '../../utils'; // Assuming these exist

interface MediaDetailModalProps {
  entry: ArchiveEntry | null;
  open: boolean;
  onClose: () => void;
}

const MediaDetailModal: React.FC<MediaDetailModalProps> = ({ entry, open, onClose }) => {
  const { i18n } = useI18n();

  if (!entry) {
    return null;
  }

  let metadata: any = {};
  try {
    if (entry.metadata) {
      metadata = JSON.parse(entry.metadata);
    }
  } catch (e) {
    console.error("Failed to parse metadata for entry:", entry.id, e);
    // Optionally, set a flag or specific error message in metadata to display in UI
    metadata.parseError = "Failed to load full details.";
  }

  const title = entry.title || metadata.title || i18n.t('untitledMedia'); // Fallback to generic "untitled"

  // Helper to format filesize
  const formatFilesize = (bytes?: number) => {
    if (bytes === undefined || bytes === null || isNaN(bytes)) return i18n.t('filesizeNotAvailable'); // i18n key for N/A
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const renderMetadataField = (labelKey: string, value?: string | number | null, link?: string) => {
    if (value === undefined || value === null || value === '') return null;
    return (
      <Box sx={{ mb: 1 }}>
        <Typography variant="subtitle2" component="span" sx={{ fontWeight: 'bold' }}>
          {i18n.t(labelKey)}:
        </Typography>{' '}
        {link ? (
          <Link href={link} target="_blank" rel="noopener noreferrer">
            {String(value)}
          </Link>
        ) : (
          <Typography variant="body2" component="span">
            {String(value)}
          </Typography>
        )}
      </Box>
    );
  };

  const renderChipArray = (labelKey: string, items?: string[] | null) => {
    if (!items || items.length === 0) return null;
    return (
      <Box sx={{ mb: 1 }}>
        <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 0.5 }}>
          {i18n.t(labelKey)}:
        </Typography>
        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
          {items.map((item, index) => (
            <Chip key={index} label={item} size="small" />
          ))}
        </Box>
      </Box>
    );
  };


  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth scroll="paper">
      <DialogTitle sx={{ m: 0, p: 2 }}>
        {title}
        <IconButton
          aria-label="close"
          onClick={onClose}
          sx={{ position: 'absolute', right: 8, top: 8, color: (theme) => theme.palette.grey[500] }}
        >
          <CloseIcon />
        </IconButton>
      </DialogTitle>
      <DialogContent dividers>
        {metadata.parseError && (
            <Typography color="error" sx={{mb: 2}}>{metadata.parseError}</Typography>
        )}
        <Grid container spacing={3}>
          <Grid item xs={12} md={4}>
            {metadata.thumbnail ? (
              <Box component="img" src={metadata.thumbnail} alt={title} sx={{ width: '100%', borderRadius: 1, maxHeight: 300, objectFit: 'contain' }} />
            ) : (
              <Box sx={{ width: '100%', height: 200, backgroundColor: 'grey.200', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', borderRadius: 1 }}>
                <MovieIcon sx={{ fontSize: 60, color: 'grey.500' }} />
                <Typography variant="caption">{i18n.t('noThumbnail')}</Typography>
              </Box>
            )}
          </Grid>
          <Grid item xs={12} md={8}>
            {/* Uploader/Channel */}
            {metadata.uploader && renderMetadataField(metadata.channel_url ? 'channelLabel' : 'uploaderLabel', metadata.uploader || metadata.channel, metadata.channel_url)}
            {!metadata.uploader && metadata.channel && renderMetadataField('channelLabel', metadata.channel, metadata.channel_url)}


            {/* Upload Date, Duration, Resolution */}
            {renderMetadataField('uploadDateLabel', metadata.upload_date ? formatDate(metadata.upload_date) : metadata.upload_date)}
            {renderMetadataField('durationLabel', metadata.duration ? formatDuration(metadata.duration) : null)}
            {renderMetadataField('resolutionLabel', metadata.resolution || (metadata.width && metadata.height ? `${metadata.width}x${metadata.height}` : null))}

            <Divider sx={{ my: 2 }} />

            {/* Playlist/Series Info */}
            {metadata.playlist_title && renderMetadataField('playlistInfoModalLabel', `${metadata.playlist_title}${metadata.playlist_index ? ` (Part ${metadata.playlist_index})` : ''}`)}
            {metadata.season_number && metadata.episode_number && renderMetadataField('seriesInfoModalLabel', `S${metadata.season_number} E${metadata.episode_number}`)}
            {metadata.episode_title && metadata.episode_title !== title && !metadata.season_number && renderMetadataField('episodeInfoLabel', metadata.episode_title)}


            { (metadata.playlist_title || (metadata.season_number && metadata.episode_number)) && <Divider sx={{my:2}}/> }


            {/* Description (Scrollable Box) */}
            {metadata.description && (
              <Box sx={{ mb: 2 }}>
                <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 0.5 }}>
                  {i18n.t('descriptionLabel')}:
                </Typography>
                <Box sx={{ maxHeight: 150, overflowY: 'auto', pr: 1, typography: 'body2' }} dangerouslySetInnerHTML={{ __html: metadata.description.replace(/\n/g, '<br />') }} />
              </Box>
            )}

            {/* Tags (Chip array) */}
            {renderChipArray('tagsLabel', metadata.tags)}

            {/* Categories (Chip array) */}
            {renderChipArray('categoriesLabel', metadata.categories)}

            {(metadata.tags?.length > 0 || metadata.categories?.length > 0) && <Divider sx={{my:2}}/>}

            {/* Video Format Details section */}
            <Typography variant="h6" gutterBottom sx={{mt:1}}>{i18n.t('technicalDetailsLabel')}</Typography> {/* New i18n key */}
            {renderMetadataField('formatExtLabel', metadata.ext)}
            {renderMetadataField('formatVCodecLabel', metadata.vcodec)}
            {renderMetadataField('formatACodecLabel', metadata.acodec)}
            {renderMetadataField('formatNoteLabel', metadata.format_note)}
            {renderMetadataField('filesizeLabel', formatFilesize(metadata.filesize_approx))}


            <Divider sx={{ my: 2 }} />

            {/* Source URL */}
            {renderMetadataField('sourceUrlLabel', entry.source, entry.source)}

          </Grid>
        </Grid>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{i18n.t('closeButton')}</Button>
      </DialogActions>
    </Dialog>
  );
};

export default MediaDetailModal;

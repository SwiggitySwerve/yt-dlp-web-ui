import {
  Checkbox,
  Container,
  FormControl,
  FormControlLabel,
  Grid,
  InputAdornment,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  SelectChangeEvent,
  Slider,
  Stack,
  Switch,
  TextField,
  Typography,
  capitalize,
  Box, // Added
  List, // Added
  ListItem, // Added
  ListItemText, // Added
  IconButton, // Added
  Button // Added
} from '@mui/material'
import { useAtom } from 'jotai'
import { Suspense, useCallback, useEffect, useMemo, useState } from 'react'
import {
  Subject,
  debounceTime,
  distinctUntilChanged,
  map,
  takeWhile
} from 'rxjs'
import { rpcPollingTimeState } from '../atoms/rpc'
import {
  Accent,
  Language,
  Theme,
  accentState,
  accents,
  appTitleState,
  enableCustomArgsState,
  fileRenamingState,
  autoFileExtensionState,
  formatSelectionState,
  languageState,
  languages,
  pathOverridingState,
  servedFromReverseProxyState,
  servedFromReverseProxySubDirState,
  serverAddressState,
  serverPortState,
  themeState,
  // Added imports for preference atoms and type
  preferredFormatsAtom,
  preferredQualitiesAtom,
  PreferenceItem
} from '../atoms/settings'
import CookiesTextField from '../components/CookiesTextField'
import UpdateBinaryButton from '../components/UpdateBinaryButton'
import { useToast } from '../hooks/toast'
import { useI18n } from '../hooks/useI18n'
import Translator from '../lib/i18n'
import { validateDomain, validateIP } from '../utils'
// Added imports for Icons
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import { v4 as uuidv4 } from 'uuid'; // For generating unique IDs


// NEED ABSOLUTELY TO BE SPLIT IN MULTIPLE COMPONENTS
export default function Settings() {
  const [reverseProxy, setReverseProxy] = useAtom(servedFromReverseProxyState)
  const [baseURL, setBaseURL] = useAtom(servedFromReverseProxySubDirState)

  const [formatSelection, setFormatSelection] = useAtom(formatSelectionState)
  const [pathOverriding, setPathOverriding] = useAtom(pathOverridingState)
  const [fileRenaming, setFileRenaming] = useAtom(fileRenamingState)
  const [autoFileExtension, setAutoFileExtension] = useAtom(autoFileExtensionState)
  const [enableArgs, setEnableArgs] = useAtom(enableCustomArgsState)

  const [serverAddr, setServerAddr] = useAtom(serverAddressState)
  const [serverPort, setServerPort] = useAtom(serverPortState)

  const [pollingTime, setPollingTime] = useAtom(rpcPollingTimeState)
  const [language, setLanguage] = useAtom(languageState)
  const [appTitle, setApptitle] = useAtom(appTitleState)
  const [accent, setAccent] = useAtom(accentState)

  const [theme, setTheme] = useAtom(themeState)

  // Added atoms for preferences
  const [formats, setFormats] = useAtom(preferredFormatsAtom);
  const [qualities, setQualities] = useAtom(preferredQualitiesAtom);

  const [invalidIP, setInvalidIP] = useState(false)

  const { i18n } = useI18n()

  const { pushMessage } = useToast()

  const baseURL$ = useMemo(() => new Subject<string>(), [])
  const serverAddr$ = useMemo(() => new Subject<string>(), [])
  const serverPort$ = useMemo(() => new Subject<string>(), [])

  const [, updateState] = useState({})
  const forceUpdate = useCallback(() => updateState({}), [])

  useEffect(() => {
    const sub = baseURL$
      .pipe(debounceTime(500))
      .subscribe(baseURL => {
        setBaseURL(baseURL)
        pushMessage(i18n.t('restartAppMessage'), 'info')
      })
    return () => sub.unsubscribe()
  }, []) // Removed baseURL$ from deps as it's stable

  useEffect(() => {
    const sub = serverAddr$
      .pipe(
        debounceTime(500),
        distinctUntilChanged()
      )
      .subscribe(addr => {
        if (validateIP(addr)) {
          setInvalidIP(false)
          setServerAddr(addr)
          pushMessage(i18n.t('restartAppMessage'), 'info')
        } else if (validateDomain(addr)) {
          setInvalidIP(false)
          setServerAddr(addr)
          pushMessage(i18n.t('restartAppMessage'), 'info')
        } else {
          setInvalidIP(true)
        }
      })
    return () => sub.unsubscribe()
  }, [serverAddr$, setServerAddr, pushMessage, i18n]) // Added missing deps

  useEffect(() => {
    const sub = serverPort$
      .pipe(
        debounceTime(500),
        map(val => Number(val)),
        takeWhile(val => isFinite(val) && val <= 65535),
      )
      .subscribe(port => {
        setServerPort(port)
        pushMessage(i18n.t('restartAppMessage'), 'info')
      })
    return () => sub.unsubscribe()
  }, [serverPort$, setServerPort, pushMessage, i18n]) // Added missing deps

  /**
   * Language toggler handler 
   */
  const handleLanguageChange = (event: SelectChangeEvent<Language>) => {
    setLanguage(event.target.value as Language)

    Translator.instance.setLanguage(event.target.value)
    setTimeout(() => {
      forceUpdate()
    }, 100)
  }

  /**
   * Theme toggler handler 
   */
  const handleThemeChange = (event: SelectChangeEvent<Theme>) => {
    setTheme(event.target.value as Theme)
  }

  return (
    <Container maxWidth="xl" sx={{ mt: 4, mb: 8 }}>
      <Paper
        sx={{
          p: 2.5,
          display: 'flex',
          flexDirection: 'column',
          minHeight: 240,
        }}
      >
        {/* Existing settings UI */}
        <Typography pb={2} variant="h6" color="primary">
          {i18n.t('settingsAnchor')}
        </Typography>
        <Grid container spacing={2}>
          {/* Server Address, Port, App Title, Polling Time, Reverse Proxy */}
          <Grid item xs={12} md={11}>
            <TextField
              fullWidth
              label={i18n.t('serverAddressTitle')}
              defaultValue={serverAddr}
              error={invalidIP}
              onChange={(e) => serverAddr$.next(e.currentTarget.value)}
              InputProps={{
                startAdornment: <InputAdornment position="start">ws://</InputAdornment>,
              }}
            />
          </Grid>
          <Grid item xs={12} md={1}>
            <TextField
              disabled={reverseProxy}
              fullWidth
              label={i18n.t('serverPortTitle')}
              defaultValue={serverPort}
              onChange={(e) => serverPort$.next(e.currentTarget.value)}
              error={isNaN(Number(serverPort)) || Number(serverPort) > 65535}
            />
          </Grid>
          <Grid item xs={12} md={12}>
            <TextField
              fullWidth
              label={i18n.t('appTitle')}
              defaultValue={appTitle}
              onChange={(e) => setApptitle(e.currentTarget.value)}
              error={appTitle === ''}
            />
          </Grid>
          <Grid item xs={12} md={12}>
            <Typography>
              {i18n.t('rpcPollingTimeTitle')}
            </Typography>
            <Typography variant='caption' sx={{ mb: 0.5 }}>
              {i18n.t('rpcPollingTimeDescription')}
            </Typography>
            <Slider
              aria-label="rpc polling time"
              defaultValue={pollingTime}
              max={2000}
              getAriaValueText={(v: number) => `${v} ms`}
              step={null}
              valueLabelDisplay="off"
              marks={[
                { value: 100, label: '100 ms' },
                { value: 250, label: '250 ms' },
                { value: 500, label: '500 ms' },
                { value: 750, label: '750 ms' },
                { value: 1000, label: '1000 ms' },
                { value: 2000, label: '2000 ms' },
              ]}
              onChange={(_, value) => typeof value === 'number'
                ? setPollingTime(value)
                : setPollingTime(1000)
              }
            />
          </Grid>
          <Grid item xs={12}>
            <Typography variant="h6" color="primary" sx={{ mb: 0.5 }}>
              Reverse Proxy
            </Typography>
            <FormControlLabel
              control={
                <Checkbox
                  defaultChecked={reverseProxy}
                  onChange={() => setReverseProxy(state => !state)}
                />
              }
              label={i18n.t('servedFromReverseProxyCheckbox')}
              sx={{ mb: 1 }}
            />
            <TextField
              fullWidth
              label={i18n.t('urlBase')}
              defaultValue={baseURL}
              onChange={(e) => {
                let value = e.currentTarget.value
                if (value.startsWith('/')) {
                  value = value.substring(1)
                }
                if (value.endsWith('/')) {
                  value = value.substring(0, value.length - 1)
                }
                baseURL$.next(value)
              }}
              sx={{ mb: 2 }}
            />
          </Grid>
        </Grid>
        
        {/* Appearance Settings */}
        <Typography variant="h6" color="primary" sx={{ mt: 0.5, mb: 2 }}>
          Appearance
        </Typography>
        <Grid container spacing={2}>
          {/* Language, Theme, Accent */}
          <Grid item xs={12}>
            <FormControl fullWidth>
              <InputLabel>{i18n.t('languageSelect')}</InputLabel>
              <Select
                defaultValue={language}
                label={i18n.t('languageSelect')}
                onChange={handleLanguageChange}
              >
                {languages.toSorted((a, b) => a.localeCompare(b)).map(l => (
                  <MenuItem value={l} key={l}>
                    {capitalize(l)}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={6}>
            <FormControl fullWidth>
              <InputLabel>{i18n.t('themeSelect')}</InputLabel>
              <Select
                defaultValue={theme}
                label={i18n.t('themeSelect')}
                onChange={handleThemeChange}
              >
                <MenuItem value="light">{i18n.t('lightThemeButton')}</MenuItem>
                <MenuItem value="dark">{i18n.t('darkThemeButton')}</MenuItem>
                <MenuItem value="system">System</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={6}>
            <FormControl fullWidth>
              <InputLabel>{i18n.t('accentSelect')}</InputLabel>
              <Select
                defaultValue={accent}
                label={i18n.t('accentSelect')}
                onChange={(e) => setAccent(e.target.value as Accent)}
              >
                {accents.map((accent) => (
                  <MenuItem key={accent} value={accent}>
                    {capitalize(accent)}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
        </Grid>

        {/* Download Preferences UI */}
        <Paper sx={{ p: 2, mt: 3, mb: 3 }}> {/* Grouping preferences */}
          <Typography variant="h6" gutterBottom>{i18n.t('preferredFormatsTitle')}</Typography>
          <List dense>
            {formats.map((item) => (
              <ListItem
                key={item.id}
                secondaryAction={
                  <IconButton edge="end" aria-label="delete" onClick={() => {
                    setFormats(formats.filter(f => f.id !== item.id));
                  }}>
                    <DeleteIcon />
                  </IconButton>
                }
              >
                <Checkbox
                  edge="start"
                  checked={item.enabled}
                  onChange={(e) => {
                    setFormats(formats.map(f => f.id === item.id ? { ...f, enabled: e.target.checked } : f));
                  }}
                />
                <ListItemText primary={item.value} />
              </ListItem>
            ))}
          </List>
          <Box sx={{ display: 'flex', alignItems: 'center', mt: 1, gap: 1 }}>
            <TextField 
                id="new-format-input"
                size="small" 
                label={i18n.t('addNewFormatLabel')}
                variant="outlined" 
            />
            <Button 
                variant="contained" 
                startIcon={<AddIcon />}
                onClick={() => {
                    const inputElement = document.getElementById('new-format-input') as HTMLInputElement;
                    const newValue = inputElement?.value.trim();
                    if (newValue) {
                        setFormats([...formats, { id: uuidv4(), value: newValue, enabled: true }]);
                        if (inputElement) inputElement.value = ''; 
                    }
                }}
            >
                {i18n.t('addButtonLabel')}
            </Button>
          </Box>
        </Paper>

        <Paper sx={{ p: 2, mb: 3 }}>
          <Typography variant="h6" gutterBottom>{i18n.t('preferredQualitiesTitle')}</Typography>
          <List dense>
            {qualities.map((item) => (
              <ListItem
                key={item.id}
                secondaryAction={
                  <IconButton edge="end" aria-label="delete" onClick={() => {
                    setQualities(qualities.filter(q => q.id !== item.id));
                  }}>
                    <DeleteIcon />
                  </IconButton>
                }
              >
                <Checkbox
                  edge="start"
                  checked={item.enabled}
                  onChange={(e) => {
                    setQualities(qualities.map(q => q.id === item.id ? { ...q, enabled: e.target.checked } : q));
                  }}
                />
                <ListItemText primary={item.value} />
              </ListItem>
            ))}
          </List>
          <Box sx={{ display: 'flex', alignItems: 'center', mt: 1, gap: 1 }}>
            <TextField 
                id="new-quality-input"
                size="small" 
                label={i18n.t('addNewQualityLabel')}
                variant="outlined" 
            />
            <Button 
                variant="contained" 
                startIcon={<AddIcon />}
                onClick={() => {
                    const inputElement = document.getElementById('new-quality-input') as HTMLInputElement;
                    const newValue = inputElement?.value.trim();
                    if (newValue) {
                        setQualities([...qualities, { id: uuidv4(), value: newValue, enabled: true }]);
                        if (inputElement) inputElement.value = '';
                    }
                }}
            >
                {i18n.t('addButtonLabel')}
            </Button>
          </Box>
        </Paper>
        {/* End Download Preferences UI */}
        
        {/* General Download Settings */}
        <Typography variant="h6" color="primary" sx={{ mt: 2, mb: 0.5 }}>
          {i18n.t('generalDownloadSettings')}
        </Typography>
        <FormControlLabel
          control={
            <Switch
              defaultChecked={formatSelection}
              onChange={() => {
                setFormatSelection(!formatSelection)
              }}
            />
          }
          label={i18n.t('formatSelectionEnabler')}
        />
        <Grid>
          <Typography variant="h6" color="primary" sx={{ mt: 2, mb: 0.5 }}>
            {i18n.t('overridesAnchor')}
          </Typography>
          <Stack direction="column">
            {/* Path Overriding, File Renaming, Auto File Extension, Custom Args Switches */}
            <FormControlLabel
              control={
                <Switch
                  defaultChecked={!!pathOverriding}
                  onChange={() => {
                    setPathOverriding(state => !state)
                  }}
                />
              }
              label={i18n.t('pathOverrideOption')}
            />
            <FormControlLabel
              control={
                <Switch
                  defaultChecked={fileRenaming}
                  onChange={() => {
                    if (fileRenaming) {
                      setAutoFileExtension(false)
                    }
                    setFileRenaming(state => !state)
                  }}
                />
              }
              label={i18n.t('filenameOverrideOption')}
            />
            {
              <FormControlLabel
                control={
                  <Switch
                    disabled={!fileRenaming}
                    checked={fileRenaming ? autoFileExtension : false}
                    defaultChecked={autoFileExtension}
                    onChange={() => {
                      setAutoFileExtension(state => !state)
                    }}
                  />
                }
                label={i18n.t('autoFileExtensionOption')}
              /> 
            }
            <FormControlLabel
              control={
                <Switch
                  defaultChecked={enableArgs}
                  onChange={() => {
                    setEnableArgs(state => !state)
                  }}
                />
              }
              label={i18n.t('customArgs')}
            />
          </Stack>
        </Grid>
        <Grid sx={{ mr: 1, mt: 2 }}>
          <Typography variant="h6" color="primary" sx={{ mb: 2 }}>
            Cookies
          </Typography>
          <Suspense>
            <CookiesTextField />
          </Suspense>
        </Grid>
        <Grid>
          <Stack direction="row" sx={{ pt: 2 }}>
            <UpdateBinaryButton />
          </Stack>
        </Grid>
      </Paper>
    </Container>
  )
}

# HDF5 Agent - Visual Interface Guide

## Application UI Layout

The HDF5 Agent provides a professional scientific interface for manipulating HDF5 files. Here's what you'll see:

```
╔════════════════════════════════════════════════════════════════════════════╗
║                         🔍 HDF5 Agent                                      ║
║                   Data Manipulation Interface                              ║
║                                                       [+ New File] [Help]   ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Files               │ Structure Tree      │ Data Viewer & Editor           ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  📄 test_data.h5     │ 📄 test_data.h5     │ measurements > waveform       ║
║  📄 dataset.h5       │ 📁+ measurements    │ Shape: [1000]                 ║
║  📄 analysis.h5      │    📊 waveform      │ Type: float64                 ║
║  📄 output.h5        │    📊 matrix        │ Size: 1000                    ║
║                      │    📝 labels        │                                ║
║                      │                     │ ┌─────┬─────┬─────┬──────┐   ║
║  Size: 256 KB        │                     │ │0.00 │0.01 │0.02 │0.03  │   ║
║                      │                     │ ├─────┼─────┼─────┼──────┤   ║
║                      │                     │ │0.04 │0.05 │0.06 │0.07  │   ║
║                      │                     │ ├─────┼─────┼─────┼──────┤   ║
║                      │                     │ │0.08 │0.09 │0.10 │0.11  │   ║
║                      │                     │ ├─────┼─────┼─────┼──────┤   ║
║                      │                     │ │...more...          ... more║
║                      │                     │ └─────┴─────┴─────┴──────┘   ║
║                      │                     │                                ║
║                      │                     │ Click a cell to edit 📝        ║
║                      │                     │ Press Enter to save            ║
║                      │                     │                                ║
╚════════════════════════════════════════════════════════════════════════════╝
```

## Color Scheme

### Dark Theme (Default)
```
┌────────────────────────────────┐
│ Background Primary #0f1419      │ Dark charcoal - main workspace
│ Background Secondary #1a202c    │ Slightly lighter - sidebars
│ Background Tertiary #2d3748     │ For hover states
│ Text Primary #e2e8f0            │ Main content text
│ Text Muted #94a3b8              │ Secondary labels
│ Accent Primary #00d9ff          │ Cyan - data & highlights
│ Accent Secondary #fbbf24        │ Amber - metadata
│ Error #ef4444                   │ Red - destructive
│ Success #10b981                 │ Green - confirmations
└────────────────────────────────┘
```

## Component Behaviors

### 1. File Browser (Left Sidebar)

```
Files
───────────────────────────────
[📄 test_data.h5]  ← Active/Selected
   Size: 256 KB
   
[📄 dataset.h5]
   Size: 128 KB
   
[📄] analysis.h5
   Size: 512 KB
   
[+ Create New]
```

**Interactions:**
- Click file → Load structure
- Hover → Show size
- Right-click → Delete option
- Drag → Reorder (future)

### 2. Structure Tree (Middle)

```
Structure
───────────────────────────────
▼ test_data.h5
  ▼ measurements
    📊 waveform        [1000]
    📊 matrix          [10×20]
    📝 labels          [4]
  
  (Click to expand/collapse groups)
  (Click datasets to view data)
```

**Features:**
- Hierarchical navigation
- Type icons
- Shape indicators
- Smooth transitions

### 3. Data Viewer (Right)

```
measurements > waveform                 [↓ Export]
Shape: [1000]  Type: float64  Size: 1000

┌──────────┬──────────┬──────────┬────────┐
│  0.0000  │  0.0100  │  0.0200  │ 0.0300 │
├──────────┼──────────┼──────────┼────────┤
│  0.0400  │  0.0500  │  0.0600  │ 0.0700 │
├──────────┼──────────┼──────────┼────────┤
│  0.0800  │  0.0900  │  0.1000  │ 0.1100 │
├──────────┼──────────┼──────────┼────────┤
│  ...more cells...           ....... more data
└──────────┴──────────┴──────────┴────────┘

(Click cell to edit)
(Large datasets show preview only)
```

**Features:**
- Grid layout
- Type-specific formatting
- Inline editing
- Real-time sync
- Truncation for large data

## Modal Dialogs

### Create New File

```
╔════════════════════════════════╗
║  Create New File               ║ ╳
╠════════════════════════════════╣
║                                ║
║  Filename:                     ║
║  ┌──────────────────────────┐  ║
║  │ mydata.h5                │  ║
║  └──────────────────────────┘  ║
║                                ║
╠════════════════════════════════╣
║        [Cancel]  [Create]      ║
╚════════════════════════════════╝
```

## Data Cell Interactions

### Normal State
```
┌──────────┐
│  0.5000  │  ← Click to edit
└──────────┘
```

### Editing State
```
┌──────────┐
│ 0.5500   │  ← Type new value
└──────────┘
         ↓ Press Enter or click elsewhere
```

### After Edit
```
┌──────────┐
│  0.5500  │  ← Updated & synced to file
└──────────┘
```

## Status Indicators

### Loading
```
⟳ Loading data...
```

### Success
```
✓ Data updated successfully
```

### Error
```
✗ Failed to update: Invalid value
```

### Health Status
```
● Connected (backend running)
● API responding
● Files accessible
```

## Keyboard Shortcuts

```
Key             Action
────────────────────────────────
Enter           Save cell edit
Escape          Cancel edit
Ctrl+S          Save current dataset
Ctrl+N          New file
Ctrl+O          Open file
Ctrl+/          Toggle sidebar
F5              Refresh
?               Show help
```

## Responsive Design

### Desktop (1200px+)
```
┌─ Sidebar ─┬─ Tree ─┬─── Data Viewer ───┐
│           │        │                   │
│  280px    │ 35%    │      65%           │
└───────────┴────────┴───────────────────┘
```

### Tablet (768px-1024px)
```
┌─ Sidebar ─┬─ Tree ─┬─── Data Viewer ───┐
│           │ 40%    │      40%           │
│ 20%       │        │                   │
└───────────┴────────┴───────────────────┘
```

### Mobile (< 768px)
```
┌─ Sidebar ─┐
│ 100%      │
├───────────┤
│ Tree      │
│ 100%      │
├───────────┤
│ Data      │
│ 100%      │
└───────────┘
```

## Animation Effects

### Page Load
```
1. Header slides down (150ms)
2. Sidebar fades in (250ms)  
3. Main content reveals (350ms)
4. Data cells stagger (100ms each)
```

### File Selection
```
1. File highlight (fast 150ms)
2. Tree structure loads (250ms)
3. Data loads on demand (350ms)
```

### Cell Edit
```
1. Cell becomes active (100ms)
2. Input field shows (150ms)
3. Value updates (instant)
4. Cell returns to normal (200ms)
```

## Dark/Light Theme Toggle

```
Current: Dark
───────────────────────────────
● Dark Mode (Scientific)
○ Light Mode (Clean)

Colors automatically adjust:
- Background: #0f1419 ↔ #ffffff
- Text: #e2e8f0 ↔ #1f2937
- Accent: #00d9ff (stays consistent)
```

## Accessibility Features

✓ High contrast ratios (WCAG AA)
✓ Keyboard navigation
✓ Screen reader support
✓ Focus indicators
✓ Color-blind friendly palette
✓ Large touch targets (44x44px minimum)
✓ Readable fonts (13-16px)

## Example Workflow

### Step 1: Open File
```
User clicks "test_data.h5"
    ↓
Sidebar highlights file
Tree builds structure
```

### Step 2: Explore Structure  
```
User expands "measurements" group
    ↓
Sees: waveform, matrix, labels datasets
```

### Step 3: View Data
```
User clicks "waveform" dataset
    ↓
Right pane shows 1000 float values
Grid displays first 100 values
```

### Step 4: Edit Value
```
User clicks cell with value 0.5000
Input field appears
User types: 0.7500
Presses Enter
    ↓
Value updates immediately
File syncs to disk
```

### Step 5: Create New File
```
User clicks "+ New File"
Modal dialog appears
User enters: "analysis.h5"
Clicks Create
    ↓
New file appears in sidebar
Ready to use
```

## Performance Indicators

### Dataset Size Warnings
```
Dataset Size        Behavior
────────────────────────────────
< 10k               Full display
10k - 100k          Preview (100 cells)
100k - 1M           Preview (50 cells)
> 1M                Summary stats only
```

### File Load Times
```
File Size     Load Time    Mode
───────────────────────────────
< 1 MB        < 100ms      Full
1-10 MB       100-500ms    Full
10-100 MB     500-2000ms   Limited
> 100 MB      Streaming    On-demand
```

## Error Display

### Connection Error
```
╔════════════════════════════════╗
║ ✗ Failed to connect to server  ║
║                                ║
║ Backend may be offline.        ║
║ Check:                         ║
║ • Is server running?           ║
║ • Port 8080 available?         ║
║                                ║
║              [Retry] [Help]    ║
╚════════════════════════════════╝
```

### Data Error
```
╔════════════════════════════════╗
║ ✗ Invalid value               ║
║                                ║
║ Expected: Float64             ║
║ Got: "text"                   ║
║                                ║
║              [Cancel] [Save]   ║
╚════════════════════════════════╝
```

---

## UI Customization

The interface uses CSS variables for easy theming:

```css
/* In App.css */
:root {
  --color-bg-primary: #0f1419;
  --color-accent-primary: #00d9ff;
  --font-display: 'IBM Plex Mono';
  /* ... more variables ... */
}
```

To customize:
1. Edit CSS variables in `App.css`
2. Rebuild: `npm run build`
3. Deploy updated frontend

---

## Browser Developer Tools

### Useful Commands
```javascript
// Check connection status
fetch('http://localhost:8080/api/health')
  .then(r => r.json())
  .then(console.log)

// List files
fetch('http://localhost:8080/api/files')
  .then(r => r.json())
  .then(console.log)

// View API schema
fetch('http://localhost:8080/api/')
  .then(r => r.text())
  .then(console.log)
```

---

This comprehensive visual guide shows the professional scientific interface that HDF5 Agent provides for intuitive HDF5 file manipulation.

**Ready to start? Run `make dev` and open http://localhost:3000!** 🎉

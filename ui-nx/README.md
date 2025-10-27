# Waiting Room Frontend - Nx Monorepo

This is the frontend monorepo for the Waiting Room application, built with Angular and managed by Nx.

## 🏗️ Architecture

### Applications
- **admin** - Administrative interface for system configuration
- **backoffice** - Back office management interface
- **kiosk** - Self-service kiosk interface
- **mobile** - Mobile-responsive interface
- **tv** - TV display interface

### Libraries
- **api-client** - HTTP client and WebSocket services for API communication
- **ui** - Shared UI components and utilities
- **primeng-components** - PrimeNG component library with Tailwind CSS integration

## 🚀 Getting Started

### Prerequisites
- Node.js 18+ 
- pnpm (recommended) or npm

### Installation
```bash
pnpm install
```

### Development
```bash
# Serve individual applications
pnpm run serve:admin
pnpm run serve:backoffice
pnpm run serve:kiosk
pnpm run serve:mobile
pnpm run serve:tv

# Build all applications
pnpm run build:all

# Build individual applications
pnpm run build:admin
pnpm run build:backoffice
pnpm run build:kiosk
pnpm run build:mobile
pnpm run build:tv
```

### Libraries
```bash
# Build all libraries
pnpm run build:api-client
pnpm run build:ui
pnpm run build:primeng-components

# Test libraries
pnpm run test:api-client
pnpm run test:ui
pnpm run test:primeng-components
```

## 🧪 Testing

```bash
# Test all projects
pnpm run test:all

# Test specific project
pnpm run test:admin
pnpm run test:backoffice
pnpm run test:kiosk
pnpm run test:mobile
pnpm run test:tv
```

## 🔍 Linting

```bash
# Lint all projects
pnpm run lint:all

# Lint specific project
pnpm run lint:admin
pnpm run lint:backoffice
pnpm run lint:kiosk
pnpm run lint:mobile
pnpm run lint:tv
```

## 🎨 Styling

This project uses:
- **Tailwind CSS** for utility-first styling
- **PrimeNG** for component library
- **PrimeIcons** for iconography

The combination of Tailwind CSS and PrimeNG provides:
- Rapid development with utility classes
- Professional components out of the box
- Consistent design system
- Easy customization

## 📦 Dependencies

### Core Dependencies
- Angular 20.3.0
- PrimeNG 20.2.0
- PrimeIcons 7.0.0
- Tailwind CSS 4.1.16
- RxJS 7.8.0

### Development Dependencies
- Nx 22.0.1
- Jest for testing
- ESLint for linting
- Prettier for formatting

## 🏛️ Project Structure

```
apps/
├── admin/           # Admin application
├── backoffice/      # Back office application
├── kiosk/          # Kiosk application
├── mobile/         # Mobile application
└── tv/             # TV application

libs/
├── api-client/     # API client library
├── ui/             # Shared UI components
└── primeng-components/ # PrimeNG components
```

## 🔧 Nx Commands

```bash
# Generate new application
npx nx g @nx/angular:app my-app

# Generate new library
npx nx g @nx/angular:lib my-lib

# Show project graph
npx nx graph

# Show affected projects
npx nx affected

# Run affected projects
npx nx affected --target=build
```

## 🚀 Deployment

Each application can be built independently:

```bash
# Build for production
pnpm run build:admin
pnpm run build:backoffice
pnpm run build:kiosk
pnpm run build:mobile
pnpm run build:tv
```

The built applications will be available in the `dist/` directory.

## 📝 Notes

- All applications share common libraries for consistency
- ESLint rules enforce proper dependency boundaries
- Each application has its own environment configuration
- PrimeNG components are styled with Tailwind CSS for consistency
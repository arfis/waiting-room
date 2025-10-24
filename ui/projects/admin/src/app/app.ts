import { Component, signal } from '@angular/core';
import { RouterOutlet, Router, NavigationEnd, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { filter } from 'rxjs/operators';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, RouterOutlet, RouterModule],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class AppComponent {
  title = 'admin';
  
  sidebarOpen = signal(true); // Start with sidebar open
  currentRoute = signal('');

  constructor(private router: Router) {
    // Track current route for page title
    this.router.events
      .pipe(filter(event => event instanceof NavigationEnd))
      .subscribe((event: NavigationEnd) => {
        this.currentRoute.set(event.url);
        // Close sidebar on mobile after navigation
        if (window.innerWidth < 1024) {
          this.sidebarOpen.set(false);
        }
      });
  }

  toggleSidebar(): void {
    this.sidebarOpen.set(!this.sidebarOpen());
  }

  closeSidebar(): void {
    this.sidebarOpen.set(false);
  }


  getPageTitle(): string {
    const route = this.currentRoute();
    switch (route) {
      case '/dashboard':
        return 'Dashboard';
      case '/configuration':
        return 'Configuration';
      case '/card-readers':
        return 'Card Readers';
      default:
        return 'Admin Panel';
    }
  }
}
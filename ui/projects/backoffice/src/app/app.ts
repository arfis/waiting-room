import { Component, signal } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { BackofficeComponent } from './backoffice.component';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, BackofficeComponent],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App {
  protected readonly title = signal('backoffice');
}

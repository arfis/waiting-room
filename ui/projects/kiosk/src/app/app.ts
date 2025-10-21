import { Component } from '@angular/core';
import { CardLoaderComponent } from './card-loader/card-loader.component';
import {RouterOutlet} from '@angular/router';

@Component({
  selector: 'app-root',
  imports: [CardLoaderComponent, RouterOutlet],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class AppComponent {
  protected readonly title = 'kiosk';
}

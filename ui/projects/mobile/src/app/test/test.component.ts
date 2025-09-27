import { Component } from '@angular/core';

@Component({
  selector: 'app-test',
  standalone: true,
  template: `
    <div class="p-4">
      <h1 class="text-2xl font-bold">Test Component</h1>
      <p>This is a simple test component to verify Angular is working.</p>
    </div>
  `
})
export class TestComponent {
}

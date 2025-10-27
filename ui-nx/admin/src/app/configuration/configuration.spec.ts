import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Configuration } from './configuration';

describe('Configuration', () => {
  let component: Configuration;
  let fixture: ComponentFixture<Configuration>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Configuration]
    })
    .compileComponents();

    fixture = TestBed.createComponent(Configuration);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

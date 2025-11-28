import { ComponentFixture, TestBed } from '@angular/core/testing';
import { PriorityConfigurationComponent } from './priority-configuration';

describe('PriorityConfigurationComponent', () => {
  let component: PriorityConfigurationComponent;
  let fixture: ComponentFixture<PriorityConfigurationComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PriorityConfigurationComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(PriorityConfigurationComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

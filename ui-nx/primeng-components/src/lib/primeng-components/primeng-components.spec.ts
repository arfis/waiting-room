import { ComponentFixture, TestBed } from '@angular/core/testing';
import { PrimengComponents } from './primeng-components';

describe('PrimengComponents', () => {
  let component: PrimengComponents;
  let fixture: ComponentFixture<PrimengComponents>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PrimengComponents],
    }).compileComponents();

    fixture = TestBed.createComponent(PrimengComponents);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

import React from 'react';
import { Plus } from 'lucide-react';
import { Project } from '../types';
import { formatINR } from './DashboardView';

interface ProjectsViewProps {
  projects: Project[];
  onOpenNewProject: () => void;
  searchQuery: string;
}

export const ProjectsView: React.FC<ProjectsViewProps> = ({
  projects,
  onOpenNewProject,
  searchQuery,
}) => {
  const filtered = projects.filter((p) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return p.name.toLowerCase().includes(q) || p.code.toLowerCase().includes(q);
  });

  return (
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '20px',
          flexWrap: 'wrap',
          gap: '12px',
        }}
      >
        <div>
          <h2 style={{ fontSize: '1.2rem', fontWeight: 700 }}>
            National Highway Infrastructure Projects
          </h2>
          <p style={{ fontSize: '0.8rem', color: 'var(--text-dim)' }}>
            Corridors, expressways, and asset development projects
          </p>
        </div>
        <button className="btn-primary" onClick={onOpenNewProject}>
          <Plus size={15} />
          <span>Add New Corridor</span>
        </button>
      </div>

      <div className="project-grid">
        {filtered.length === 0 ? (
          <div className="empty-state" style={{ gridColumn: '1 / -1' }}>
            <h4>No Corridors Found</h4>
            <p>Try clearing your search query or register a new project.</p>
          </div>
        ) : (
          filtered.map((p) => (
            <div key={p.project_id || p.code} className="project-card">
              <div className="project-code mono">{p.code || 'NH-CORRIDOR'}</div>
              <div className="project-name">{p.name}</div>
              <div className="project-meta-row">
                <span>Sanctioned Budget:</span>
                <span className="mono" style={{ fontWeight: 700, color: 'var(--emerald)' }}>
                  {formatINR(p.budget || 250000000)}
                </span>
              </div>
              <div className="project-meta-row">
                <span>Status:</span>
                <span className="badge badge-approved">{p.status || 'ACTIVE CORRIDOR'}</span>
              </div>
              <div className="project-meta-row">
                <span>Assigned Division:</span>
                <span>NHIT Trust Division</span>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

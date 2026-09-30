# Product Management Documentation

This directory contains product management documentation for **hdf5-agent**, the HTTP service for browsing and editing HDF5 files. These documents support internal planning, external communications, and stakeholder alignment.

## Document Index

### Internal Documents

| Document | Purpose | Audience | Update Frequency |
|----------|---------|----------|------------------|
| [`vision.md`](./vision.md) | Product vision and problem statement | PM, eng, leadership | Quarterly or when strategy shifts |
| [`prd.md`](./prd.md) | Product Requirements Document (MVP baseline) | PM, eng, design | Per milestone; versioned |
| [`personas-and-jtbd.md`](./personas-and-jtbd.md) | User personas and jobs-to-be-done | PM, design, marketing | Quarterly; after user research |
| [`tech-analysis-go-vs-python.md`](./tech-analysis-go-vs-python.md) | Technical analysis: Go vs Python backend | PM, eng leadership, architects | One-time; update if re-evaluation needed |
| [`roadmap.md`](./roadmap.md) | Product roadmap (Now/Next/Later) | PM, eng, leadership, stakeholders | Monthly or per planning cycle |
| [`internal-status-update.md`](./internal-status-update.md) | Template and example for stakeholder updates | PM → leadership/partners | Weekly or bi-weekly; template is stable |
| [`messaging.md`](./messaging.md) | Positioning and messaging guide | PM, marketing, sales, support | Quarterly; before launches |

### External Documents

| Document | Purpose | Audience | Update Frequency |
|----------|---------|----------|------------------|
| [`external-overview.md`](./external-overview.md) | One-pager for partners and analysts | External stakeholders, prospects | Per major release or when partnerships form |

### Optional Reference

| Document | Purpose | Audience | Update Frequency |
|----------|---------|----------|------------------|
| [`glossary.md`](./glossary.md) | HDF5 and AI data transfer terminology | Marketing analysts, new team members, partners | As needed; living reference |

## How to Use This Suite

### For Product Managers
- **Weekly**: Update `internal-status-update.md` with current progress
- **Monthly**: Review and adjust `roadmap.md` based on velocity and feedback
- **Quarterly**: Refresh `vision.md`, `personas-and-jtbd.md`, and `messaging.md`
- **Per milestone**: Version and update `prd.md` with new acceptance criteria

### For Engineering Leads
- Reference `prd.md` for acceptance criteria and success metrics
- Consult `tech-analysis-go-vs-python.md` when evaluating backend changes
- Use `roadmap.md` to align sprint planning with product horizons

### For Marketing/Partnerships
- Use `external-overview.md` as the canonical one-pager for partners
- Reference `messaging.md` for consistent positioning in communications
- Consult `personas-and-jtbd.md` to understand target user workflows

### For Leadership
- Review `internal-status-update.md` for current state and risks
- Use `vision.md` and `roadmap.md` for strategic planning
- Reference `tech-analysis-go-vs-python.md` for architecture decisions

## Keeping Docs Updated

**Ownership**: Product Manager owns this directory. Engineering contributes to technical analyses.

**Version control**: All changes via pull request. Date significant updates in the doc header or changelog.

**Single source of truth**: These docs should not duplicate `README.md`, `ARCHITECTURE.md`, or the OpenAPI contract. Link to those sources where appropriate. When technical facts conflict, trust the repository code and `api/openapi.yaml` over PM documentation.

**Stale doc policy**: If a document hasn't been reviewed in >6 months, add a banner noting last review date and requesting validation.

## Questions or Feedback

For questions about these documents or to suggest improvements, open an issue or PR in the repository.

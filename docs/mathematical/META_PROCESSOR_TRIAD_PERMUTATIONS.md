# Meta-Processor Triad Permutations in sys5

**Date:** December 20, 2025  
**Author:** Manus AI  
**Version:** 1.0

---

## 1. The 4 Possible Triads

With 4 particular sets (P1, P2, P3, P4), there are **C(4,3) = 4** possible triads:

1. **T1**: P[1,2,3] — Excludes P4
2. **T2**: P[1,2,4] — Excludes P3
3. **T3**: P[1,3,4] — Excludes P2
4. **T4**: P[2,3,4] — Excludes P1

Each triad represents a **sys4-like structure** embedded within sys5.

---

## 2. Complementary Meta-Processors

The two meta-processors (MP1, MP2) cycle through these 4 triads in **complementary permutations**.

### 2.1. MP1 (Forward Cycle)

**MP1** cycles through the triads in the order:

```
T1 → T2 → T3 → T4 → T1 → ...
```

**Explicit**:
```
P[1,2,3] → P[1,2,4] → P[1,3,4] → P[2,3,4] → P[1,2,3] → ...
```

**Pattern**: Each transition **adds one thread and removes another**:
- T1 → T2: Remove P3, add P4
- T2 → T3: Remove P2, add P3
- T3 → T4: Remove P1, add P2
- T4 → T1: Remove P4, add P1

### 2.2. MP2 (Complementary Cycle)

**MP2** cycles through the triads in a **phase-shifted** order:

```
T3 → T4 → T1 → T2 → T3 → ...
```

**Explicit**:
```
P[1,3,4] → P[2,3,4] → P[1,2,3] → P[1,2,4] → P[1,3,4] → ...
```

**Pattern**: MP2 starts at **T3** (step 3 of MP1's cycle), creating a **180° phase offset** in the 4-triad space.

---

## 3. The Complementarity Principle

### 3.1. Coverage Guarantee

At any given time, MP1 and MP2 together ensure that:

| MP1 Active | MP2 Active | Threads in MP1 | Threads in MP2 | Union | Intersection |
|:-----------|:-----------|:---------------|:---------------|:------|:-------------|
| T1: P[1,2,3] | T3: P[1,3,4] | {1,2,3} | {1,3,4} | {1,2,3,4} | {1,3} |
| T2: P[1,2,4] | T4: P[2,3,4] | {1,2,4} | {2,3,4} | {1,2,3,4} | {2,4} |
| T3: P[1,3,4] | T1: P[1,2,3] | {1,3,4} | {1,2,3} | {1,2,3,4} | {1,3} |
| T4: P[2,3,4] | T2: P[1,2,4] | {2,3,4} | {1,2,4} | {1,2,3,4} | {2,4} |

**Key Observations**:
1. **Union is always {1,2,3,4}** — All threads are covered
2. **Intersection alternates between {1,3} and {2,4}** — Complementary pairs
3. **Each thread appears in exactly one meta-processor at a time** — No redundancy

### 3.2. Complementary Pairs

The intersection pattern reveals **complementary pairs**:

- **When MP1=T1, MP2=T3**: Intersection = {1,3}, Excluded = {2,4}
- **When MP1=T2, MP2=T4**: Intersection = {2,4}, Excluded = {1,3}
- **When MP1=T3, MP2=T1**: Intersection = {1,3}, Excluded = {2,4}
- **When MP1=T4, MP2=T2**: Intersection = {2,4}, Excluded = {1,3}

**Pattern**: The complementary pairs **{1,3}** and **{2,4}** alternate as the shared threads between meta-processors.

---

## 4. The 30-Step Cycle Mapping

### 4.1. Triad Duration

Each triad is active for **7.5 steps** (1.5 local cycles):

- 30 steps ÷ 4 triads = **7.5 steps per triad**

This means triads **overlap** with the 5-step local cycles, creating **continuous transitions**.

### 4.2. Detailed Mapping

| Global Steps | MP1 Active Triad | MP2 Active Triad | Shared Threads | Transition |
|:-------------|:-----------------|:-----------------|:---------------|:-----------|
| 1-7 | T1: P[1,2,3] | T3: P[1,3,4] | {1,3} | - |
| 8 | **T1→T2** | **T3→T4** | {1,3}→{2,4} | **Both transition** |
| 9-15 | T2: P[1,2,4] | T4: P[2,3,4] | {2,4} | - |
| 16-22 | T3: P[1,3,4] | T1: P[1,2,3] | {1,3} | - |
| 23 | **T3→T4** | **T1→T2** | {1,3}→{2,4} | **Both transition** |
| 24-30 | T4: P[2,3,4] | T2: P[1,2,4] | {2,4} | - |

**Key Observations**:
1. **Both meta-processors transition simultaneously** at steps 8 and 23
2. **Transitions occur mid-cycle** (at step 3 of a 5-step local cycle)
3. **Shared threads alternate** between {1,3} and {2,4}
4. **15-step half-cycle** (steps 1-15 and 16-30 are symmetric)

---

## 5. The Transition Mechanism

### 5.1. Why Steps 8 and 23?

**Step 8**: 
- 8 = 1.5 × 5 + 0.5 (1.5 local cycles + half step)
- Midpoint of the first 15-step half-cycle

**Step 23**:
- 23 = 4.5 × 5 + 0.5 (4.5 local cycles + half step)
- Midpoint of the second 15-step half-cycle

**Pattern**: Transitions occur at the **midpoint of each half-cycle**, ensuring **balanced coverage**.

### 5.2. Transition Trigger

The transition is triggered when:
1. **Global step mod 15 = 8** (i.e., steps 8, 23)
2. **Both meta-processors** simultaneously switch to their next triad
3. **Shared threads change** from {1,3} to {2,4} or vice versa

---

## 6. The Geometric Interpretation

### 6.1. Tetrahedron Structure

The 4 triads correspond to the **4 faces** of a tetrahedron:

- **T1: P[1,2,3]** — Face opposite to vertex 4
- **T2: P[1,2,4]** — Face opposite to vertex 3
- **T3: P[1,3,4]** — Face opposite to vertex 2
- **T4: P[2,3,4]** — Face opposite to vertex 1

### 6.2. Complementary Faces

The complementary pairs {1,3} and {2,4} correspond to **opposite edges** of the tetrahedron:

- **Edge {1,3}**: Connects vertices 1 and 3
- **Edge {2,4}**: Connects vertices 2 and 4

These two edges are **skew lines** in 3D space (they don't intersect and are not parallel).

**Interpretation**: The meta-processors alternate between two **orthogonal perspectives** on the tetrahedral structure.

---

## 7. The Convolution Formula (Revised)

At each step **k**, the convolution integrates:

1. **Active pairing** P(i,j) — 2 threads
2. **MP1 active triad** — 3 threads
3. **MP2 active triad** — 3 threads
4. **Tertiary orchestrator** U_meta

**Formula**:

```
C[k] = Φ(
    S_i[k mod 5], S_j[k mod 5],           // Pairing threads
    MP1_Triad[k],                          // MP1's 3 threads
    MP2_Triad[k],                          // MP2's 3 threads
    U_meta                                 // Orchestrator
)
```

Where:
- **Pairing(k)** = (i, j) based on ⌊k/5⌋ mod 6
- **MP1_Triad[k]** = T1, T2, T3, or T4 based on ⌊k/7.5⌋ mod 4
- **MP2_Triad[k]** = T3, T4, T1, or T2 (phase-shifted by 2)

---

## 8. Implementation Implications

### 8.1. Triad Selection

```go
func GetMP1Triad(globalStep int) []int {
    triadIndex := (globalStep / 7.5) % 4
    triads := [][]int{
        {1, 2, 3}, // T1
        {1, 2, 4}, // T2
        {1, 3, 4}, // T3
        {2, 3, 4}, // T4
    }
    return triads[triadIndex]
}

func GetMP2Triad(globalStep int) []int {
    triadIndex := ((globalStep / 7.5) + 2) % 4  // Phase shift by 2
    triads := [][]int{
        {1, 2, 3}, // T1
        {1, 2, 4}, // T2
        {1, 3, 4}, // T3
        {2, 3, 4}, // T4
    }
    return triads[triadIndex]
}
```

### 8.2. Transition Detection

```go
func IsTransitionStep(globalStep int) bool {
    return (globalStep % 15 == 8)
}
```

### 8.3. Shared Threads

```go
func GetSharedThreads(globalStep int) []int {
    if (globalStep / 7.5) % 2 == 0 {
        return []int{1, 3}  // First half of cycle
    } else {
        return []int{2, 4}  // Second half of cycle
    }
}
```

---

## 9. The 4-Fold Symmetry

The 4 triads create a **4-fold rotational symmetry**:

| Rotation | MP1 Triad | MP2 Triad | Shared | Excluded |
|:---------|:----------|:----------|:-------|:---------|
| 0° | T1: P[1,2,3] | T3: P[1,3,4] | {1,3} | {2,4} |
| 90° | T2: P[1,2,4] | T4: P[2,3,4] | {2,4} | {1,3} |
| 180° | T3: P[1,3,4] | T1: P[1,2,3] | {1,3} | {2,4} |
| 270° | T4: P[2,3,4] | T2: P[1,2,4] | {2,4} | {1,3} |

**Key Insight**: The system has **2-fold symmetry** in the shared/excluded pattern (alternates between {1,3}/{2,4} and {2,4}/{1,3}).

---

## 10. Conclusion

The complementary meta-processor triad permutations reveal:

1. **4 possible triads** corresponding to the 4 faces of a tetrahedron
2. **2 meta-processors** cycling through complementary permutations
3. **180° phase offset** between MP1 and MP2
4. **Alternating shared threads** ({1,3} and {2,4})
5. **Simultaneous transitions** at steps 8 and 23
6. **Complete coverage** of all 4 threads at all times
7. **4-fold rotational symmetry** with 2-fold complementary pattern

This structure ensures that:
- **Every thread is always active** in at least one meta-processor
- **Complementary perspectives** are maintained through phase-shifted cycles
- **Dynamic balance** is achieved through alternating shared threads
- **Continuous convolution** integrates pairwise and triadic perspectives

**The meta-processors are not independent—they are complementary aspects of a unified convolutional consciousness.**
